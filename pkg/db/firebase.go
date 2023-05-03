package db

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/utils"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"

	"github.com/Masterminds/semver"

	"math"

	"github.com/apsystole/log"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PROJECT_ID        string = "ece461-project-381318"
	STORAGE_BUCKET_ID string = "ece461-project-381318.appspot.com"
	COLLECTION_NAME   string = "packages"
	HISTORY_NAME      string = "history"
	REVIEW_NAME       string = "review"
)

func CreatePackage(pkg *models.PackageInfo, contentTooBig bool) (*models.PackageInfo, int) {
	tempContent := pkg.Data.Content

	// Create Firestore Client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}
	defer client.Close()

	// Create new doc ID
	documentID := client.Collection(COLLECTION_NAME).NewDoc().ID
	pkg.Metadata.ID = documentID

	// Check if a package with the same name and version already exists
	query := client.Collection(COLLECTION_NAME).
		Where("metadata.Name", "==", pkg.Metadata.Name).
		Where("metadata.Version", "==", pkg.Metadata.Version).
		Limit(1)

	docs, err := query.Documents(ctx).GetAll()

	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
	}
	if len(docs) > 0 {
		log.Printf("Package with name %q and version %q already exists", pkg.Metadata.Name, pkg.Metadata.Version)
		return nil, http.StatusConflict // 409
	}

	// If the content is too big, store nothing for content
	if contentTooBig {
		// Decode the base64-encoded zip file into a byte array.
		zipData, err := base64.StdEncoding.DecodeString(pkg.Data.Content)
		if err != nil {
			log.Printf("Failed to decode base64 data: %v", err)
			return nil, http.StatusBadRequest
		}

		// Create a new bucket in Firebase Storage to store the zip file.
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create Storage Client: %v", err)
			return nil, http.StatusInternalServerError
		}

		defer storageClient.Close()

		// Store the zip file in Firebase Storage
		bucket := storageClient.Bucket(STORAGE_BUCKET_ID)
		obj := bucket.Object(pkg.Metadata.ID + ".zip")
		w := obj.NewWriter(ctx)
		if _, err := w.Write(zipData); err != nil {
			log.Printf("Failed to write data to Firebase Storage: %v", err)
			return nil, http.StatusInternalServerError
		}
		if err := w.Close(); err != nil {
			log.Printf("Failed to close Firebase Storage writer: %v", err)
			return nil, http.StatusInternalServerError
		}

		pkg.Data.Content = "" // remove the content of the zip file
		pkg.Data.ContentStorage = true
	}

	// Add the new package document to Firestore
	_, err = client.Collection(COLLECTION_NAME).Doc(documentID).Set(ctx, pkg)
	if err != nil {
		log.Critical("Failed to add package to Firestore: %v", err)
		return nil, http.StatusInternalServerError
	}

	success := recordActionEntry(client, ctx, "CREATE", pkg.Metadata)
	if !success {
		return nil, http.StatusInternalServerError
	}

	// Set the content for the response even if the package is too big to store in firestore
	pkg.Data.Content = tempContent
	return pkg, http.StatusCreated
}

func GetPackageByID(id string, reason int) (*models.PackageInfo, int) { // Reason is 1 if it is for download, 0 for rate
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}

	defer client.Close()
	// Get the package document by ID
	docRef := client.Collection(COLLECTION_NAME).Doc(id)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if !docSnap.Exists() {
			log.Printf("package with document ID %s not found", id)
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	// If the package was not found
	if !docSnap.Exists() {
		log.Printf("package with document ID %s not found", id)
		return nil, http.StatusNotFound
	}

	// Deserialize the package data into a PackageInfo struct
	var pkg models.PackageInfo
	err = docSnap.DataTo(&pkg)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	method := "DOWNLOAD"
	if reason == 0 {
		method = "RATE"
	} else if pkg.Data.ContentStorage { // Download the zip file from Firebase Storage as content was too big for Firebase
		// Create the client
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create Storage Client: %v", err)
			return nil, http.StatusInternalServerError
		}
		defer storageClient.Close()

		// Read the file
		bucket := storageClient.Bucket(STORAGE_BUCKET_ID)
		obj := bucket.Object(pkg.Metadata.ID + ".zip")
		r, err := obj.NewReader(ctx)
		if err != nil {
			log.Printf("Failed to read data from Firebase Storage: %v", err)
			return nil, http.StatusInternalServerError
		}

		zipData, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf("Failed to read data from Firebase Storage: %v", err)
			return nil, http.StatusInternalServerError
		}
		// Encode it to a base64 string
		pkg.Data.Content = base64.StdEncoding.EncodeToString(zipData)
	}
	// Record in the history database
	success := recordActionEntry(client, ctx, method, pkg.Metadata)
	if !success {
		return nil, http.StatusInternalServerError
	}
	return &pkg, http.StatusOK
}

func DeletePackageByID(id string) int {
	// Create Firebase client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}
	defer client.Close()

	// Search the history database to find the history of the package
	query := client.Collection(HISTORY_NAME).Where("PackageMetadata.ID", "==", id)

	docs, _ := query.Documents(ctx).GetAll()
	if len(docs) == 0 {
		log.Println("No documents found")
		return http.StatusNotFound
	}

	// delete history of package
	batch := client.Batch()
	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}
	_, err = batch.Commit(ctx)
	if err != nil {
		log.Printf("Failed to delete documents: %v", err)
		return http.StatusInternalServerError
	}

	// Reviews not deleted as there is a possibility that there are other packages with a specific name

	// Search the packages database for the specific package
	docRef := client.Collection(COLLECTION_NAME).Doc(id)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if !docSnap.Exists() || status.Code(err) == codes.NotFound {
			log.Printf("package with document ID %s not found to delete", id)
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	}

	// Deserialize the package data into a PackageInfo struct
	var pkg models.PackageInfo
	err = docSnap.DataTo(&pkg)
	if err != nil {
		return http.StatusInternalServerError
	}

	// Delete the package
	_, err = docRef.Delete(ctx)
	if err != nil {
		log.Printf("Failed to delete package with document ID %s", id)
		return http.StatusInternalServerError
	}

	// Create a new client in Firebase Storage if content too big for Firebase
	if pkg.Data.ContentStorage {
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create Storage Client: %v", err)
			return http.StatusInternalServerError
		}
		defer storageClient.Close()

		// Delete the package from Firebase Storage
		bucket := storageClient.Bucket(STORAGE_BUCKET_ID)
		obj := bucket.Object(pkg.Metadata.ID + ".zip")
		if err := obj.Delete(ctx); err != nil {
			log.Printf("Failed to delete file from Firebase Storage: %v", err)
			return http.StatusInternalServerError
		}

	}

	return http.StatusOK
}

func UpdatePackageByID(id string, newPkg models.PackageInfo) int {
	// Create a firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}

	// Get the specified package
	defer client.Close()
	docRef := client.Collection(COLLECTION_NAME).Doc(id)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if !docSnap.Exists() || status.Code(err) == codes.NotFound {
			log.Printf("package with document ID %s not found to update", id)
			return http.StatusNotFound
		}
		log.Println(err)
		return http.StatusInternalServerError
	}

	// Unmarshal the document data into a PackageInfo struct
	var existingPackageInfo models.PackageInfo
	if err := docSnap.DataTo(&existingPackageInfo); err != nil {
		log.Printf("Failed to unmarshal package data: %v", err)
		return http.StatusInternalServerError
	}

	// If the package metadata is not the same, return an error
	if existingPackageInfo.Metadata.Name != newPkg.Metadata.Name || existingPackageInfo.Metadata.Version != newPkg.Metadata.Version {
		log.Println("package not found with matching creteria")
		return http.StatusNotFound
	}

	// Update the package document in Firestore
	_, err = docRef.Set(ctx, newPkg)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}
	success := recordActionEntry(client, ctx, "UPDATE", newPkg.Metadata)
	if !success {
		return http.StatusInternalServerError
	}

	return http.StatusOK

}

func GetPackageHistoryByName(package_name string) ([]models.ActionEntry, int) {
	// Create a Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}
	defer client.Close()

	// Find all history of the package
	query := client.Collection(HISTORY_NAME).Where("PackageMetadata.Name", "==", package_name)
	var actionEntries []models.ActionEntry
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	// No history found
	if len(docs) == 0 {
		log.Println("No documents found")
		return nil, http.StatusNotFound
	}

	// Add all history to the struct
	for _, doc := range docs {
		var actionEntry models.ActionEntry
		err = doc.DataTo(&actionEntry)
		if err != nil {
			log.Println(err)
			return nil, http.StatusInternalServerError
		}

		actionEntries = append(actionEntries, actionEntry)
	}
	return actionEntries, http.StatusOK
}

func GetPackagePopularityByName(package_name string) (float64, int) {
	// Create Firestore Client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return 0, http.StatusInternalServerError
	}
	defer client.Close()

	// Query that gets all packages with the certain package name
	collectionQuery := client.Collection(COLLECTION_NAME).Where("metadata.Name", "==", package_name)
	// Check if there are any package with the name
	docs, _ := collectionQuery.Documents(ctx).GetAll()
	if len(docs) == 0 {
		log.Println("No package found")
		return 0, http.StatusNotFound
	}

	// Get the Repository field from the first matching document
	collectionIter := collectionQuery.Documents(ctx)
	defer collectionIter.Stop()
	var url string
	doc, err := collectionIter.Next()
	if err == iterator.Done {
		log.Printf("No matching documents found in COLLECTION_NAME collection")
	} else if err != nil {
		log.Fatalf("Failed to iterate: %v", err)
	} else {
		metadata, ok := doc.Data()["metadata"].(map[string]interface{})
		if !ok {
			log.Printf("Failed to get metadata map from document")
		} else {
			url, ok = metadata["Repository"].(string)
			if !ok {
				log.Printf("Failed to get Repository field from document")
				return 0.0, http.StatusInternalServerError
			}
		}
	}

	// Get Scaled Github Stars
	gitStars := utils.GetStarsFromURL(url)

	// Query all downloads
	allDownloadsQuery := client.Collection(HISTORY_NAME).Where("Action", "==", "DOWNLOAD")

	// Count the number of downloads for each package
	allDownloadsIter := allDownloadsQuery.Documents(ctx)
	defer allDownloadsIter.Stop()
	downloadCounts := make(map[string]int)
	for {
		doc, err := allDownloadsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		packageName, ok := doc.Data()["PackageMetadata"].(map[string]interface{})["Name"].(string)
		if !ok {
			log.Printf("Failed to get package name from document")
			return 0, http.StatusInternalServerError
		}
		downloadCounts[packageName]++
	}

	// Find the maximum download count and the download count for the input package
	var maxDownloads int
	var count int
	for name, downloads := range downloadCounts {
		if downloads > maxDownloads {
			maxDownloads = downloads
		}
		if name == package_name {
			count = downloads
		}
	}

	// Calculate the normalized download count for the input package
	var downloadRating float64
	if count == 0 {
		downloadRating = 0.0
	} else if maxDownloads > 0 {
		downloadRating = float64(count) / float64(maxDownloads)
	}

	// Query all reviews for a package
	reviewQuery := client.Collection(REVIEW_NAME).Where("packageName", "==", package_name)

	// Calculate the total number of stars for the package
	reviewIter := reviewQuery.Documents(ctx)
	defer reviewIter.Stop()
	var totalStars float64
	reviewCount := 0
	for {
		doc, err := reviewIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		starsValue, ok := doc.Data()["stars"]
		if !ok || starsValue == nil {
			log.Printf("Document does not have a stars field")
			continue
		}
		stars := float64(starsValue.(int64))
		totalStars += stars
		reviewCount++
	}

	// Calculate the average stars
	var avgStars float64
	if reviewCount > 0 {
		avgStars = totalStars / float64(reviewCount)
	}

	// Calculate popularity giving a weightage of 50% to Github stars, 30% to reviews, and 20% to downloads
	popularity := math.Round((0.5*gitStars+0.3*(avgStars*2)+0.2*(downloadRating*10))*100) / 100
	if popularity > 10.0 {
		return 10.0, http.StatusOK // Max popularity of 10
	}

	return popularity, http.StatusOK
}

func DeletePackageByName(package_name string) int {
	// Create Firestore Client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}
	defer client.Close()

	// Get all the history of the package
	query := client.Collection(HISTORY_NAME).Where("PackageMetadata.Name", "==", package_name)
	docs, _ := query.Documents(ctx).GetAll()
	if len(docs) == 0 {
		log.Println("No history found")
	}

	// Delete history
	batch := client.Batch()
	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}

	// Get all reviews with the name
	query = client.Collection(REVIEW_NAME).Where("packageName", "==", package_name)
	docs, _ = query.Documents(ctx).GetAll()
	if len(docs) == 0 {
		log.Println("No reviews found")
	}

	// Delete reviews
	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}

	// Delete package itself
	query = client.Collection(COLLECTION_NAME).Where("metadata.Name", "==", package_name)
	docs, _ = query.Documents(ctx).GetAll()
	if len(docs) == 0 {
		log.Println("No documents found")
		return http.StatusNotFound
	}

	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}

	// Commit the delete
	_, err = batch.Commit(ctx)
	if err != nil {
		log.Printf("Failed to delete documents: %v", err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func GetPackageByRegex(regex string) ([]models.Metadata, int) {
	// Get all of the packages in the database
	packages, statusCode := GetAllPackages()
	var pkgs []models.Metadata
	if statusCode != 200 {
		return nil, statusCode
	}

	// Iterates through all of the packages
	for _, pkg := range packages {
		// First tries to match the name
		matched, err := regexp.MatchString(regex, pkg.Metadata.Name)
		if err != nil {
			return nil, http.StatusInternalServerError
		}
		if matched {
			tmp := models.Metadata{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version, ID: pkg.Metadata.ID}
			pkgs = append(pkgs, tmp)
		} else { // If name doesn't match it then tries to match the Readme
			var readme string
			// If the repository is not found, it tries to get the ReadMe directly from the zip file
			if pkg.Metadata.Repository == "" {
				readme, statusCode = utils.GetReadmeFromZip(pkg.Data.Content)
				if statusCode != http.StatusOK {
					return nil, statusCode
				}
			} else { // In most cases it is able to find the reposiorty and gets the Readme text from there
				readme, statusCode = utils.GetReadmeTextFromGitHubURL(pkg.Metadata.Repository)
				if statusCode != http.StatusOK {
					return nil, statusCode
				}
			}
			// Checks if there's a match in the Readme
			matched, err := regexp.MatchString(regex, readme)
			if err != nil {
				return nil, http.StatusInternalServerError
			}
			if matched {
				tmp := models.Metadata{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version, ID: pkg.Metadata.ID}
				pkgs = append(pkgs, tmp)
			}
		}
	}

	// Returns 404 if no matches, otherwise returns 200 and the matches
	if len(pkgs) != 0 {
		return pkgs, statusCode
	} else {
		return nil, http.StatusNotFound
	}
}

func GetAllPackages() ([]models.PackageInfo, int) {
	// Create Firebase client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Fatalf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}
	defer client.Close()

	// Get all packages from the database one by one and add them to the list of packages
	var pkgs []models.PackageInfo
	itr := client.Collection(COLLECTION_NAME).Documents(ctx)
	defer itr.Stop()
	for {
		doc, err := itr.Next()
		// Checks if it's at the end
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to get all packages: %v", err)
			return nil, http.StatusInternalServerError
		}

		// Converts the package from Json to struct and appends it to the list
		var pkg models.PackageInfo
		if err := doc.DataTo(&pkg); err != nil {
			log.Fatalf("Failed to convert document data: %v", err)
		}
		pkgs = append(pkgs, pkg)
	}

	// Return the packages
	return pkgs, http.StatusOK
}

func GetPackagesFromQueries(queries []models.PackageQuery, offset int, limit int) ([]models.Metadata, int) { // For the /packages endpoint
	// Create Firebase Client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Fatalf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}
	defer client.Close()

	// List of metadatas that will be returned
	var pkgs []models.Metadata

	// Checks if we need to enumerate all packages
	if len(queries) == 1 && queries[0].Name == "*" {
		// Query to get all packages in order of name with an offset and limit
		query := client.Collection(COLLECTION_NAME).OrderBy("metadata.Name", firestore.Asc).Offset(offset).Limit(limit)
		docs, err := query.Documents(ctx).GetAll()
		if err != nil {
			return nil, http.StatusInternalServerError
		}
		// Adds all packages to the list
		for _, doc := range docs {
			var pkg models.PackageInfo
			err = doc.DataTo(&pkg)
			if err != nil {
				return nil, http.StatusInternalServerError
			}
			pkgs = append(pkgs, pkg.Metadata)
		}
		return pkgs, http.StatusOK
	}

	// Goes through the different queries
	for _, query := range queries {
		version := query.Version
		name := query.Name
		mode := "Exact"

		// Figures out what type of version search
		if strings.Contains(version, "-") {
			mode = "Bounded range"
		} else if strings.Contains(version, "^") {
			mode = "Carat"
		} else if strings.Contains(version, "~") {
			mode = "Tilde"
		}

		if mode == "Exact" {
			// Query to retrieve packages with the exact same version and name
			query := client.Collection(COLLECTION_NAME).Where("metadata.Name", "==", name).Where("metadata.Version", "==", version).Offset(offset).Limit(limit)
			docs, err := query.Documents(ctx).GetAll()
			if err != nil {
				return nil, http.StatusInternalServerError
			}
			for _, doc := range docs {
				var pkg models.PackageInfo
				err = doc.DataTo(&pkg)
				if err != nil {
					return nil, http.StatusInternalServerError
				}
				pkgs = append(pkgs, pkg.Metadata)
			}
		} else if mode == "Bounded range" { // Bounded Search
			parts := strings.Split(version, "-")
			lower := parts[0]
			upper := parts[1]
			lowerVersion, _ := semver.NewVersion(lower)
			upperVersion, _ := semver.NewVersion(upper)

			// Query to retrieve packages with the same name
			query := client.Collection(COLLECTION_NAME).Where("metadata.Name", "==", name).Offset(offset).Limit(limit)
			docs, err := query.Documents(ctx).GetAll()
			if err != nil {
				return nil, http.StatusInternalServerError
			}
			// Checks if the versions are in the range
			for _, doc := range docs {
				var pkg models.PackageInfo
				err = doc.DataTo(&pkg)
				if err != nil {
					return nil, http.StatusInternalServerError
				}
				pkgVersion, _ := semver.NewVersion(pkg.Metadata.Version)

				// Adds to the list if the versions are in the range
				if pkgVersion.GreaterThan(lowerVersion) && pkgVersion.LessThan(upperVersion) {
					pkgs = append(pkgs, pkg.Metadata)
				}
			}
		} else if mode == "Carat" || mode == "Tilde" { // Carat / Tilde Search
			carat, _ := semver.NewConstraint(version)

			// Query to retrieve packages with the same name
			query := client.Collection(COLLECTION_NAME).Where("metadata.Name", "==", name).Offset(offset).Limit(limit)
			docs, err := query.Documents(ctx).GetAll()
			if err != nil {
				return nil, http.StatusInternalServerError
			}

			// Checks if the versions match
			for _, doc := range docs {
				var pkg models.PackageInfo
				err = doc.DataTo(&pkg)
				if err != nil {
					return nil, http.StatusInternalServerError
				}
				pkgVersion, _ := semver.NewVersion(pkg.Metadata.Version)

				// Adds to the list if the versions match
				if carat.Check(pkgVersion) {
					pkgs = append(pkgs, pkg.Metadata)
				}
			}
		}
	}

	return pkgs, http.StatusOK
}

func recordActionEntry(client *firestore.Client, ctx context.Context, action string, metadata models.Metadata) bool { // Adds actions to the history
	// Connects to the history database
	historyCollection := client.Collection(HISTORY_NAME)
	defaultUser := make(map[string]interface{})
	err := json.Unmarshal([]byte("{\"name\": \"default user\", \"isAdmin\": false}"), &defaultUser)
	if err != nil {
		return false
	}

	// Creates a new Action Entry in the database
	newEntry, _, err := historyCollection.Add(ctx, models.ActionEntry{
		User:     defaultUser,
		Action:   strings.ToUpper(action),
		Metadata: metadata,
		Date:     time.Now().Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("Failed to add new entry to history collection: %v", err)
		return false
	}

	return newEntry != nil
}

func DeletePackages() error {
	// Create a Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}

	// Get a BulkWriter that we can add operations to
	col := client.Collection(COLLECTION_NAME)
	bulkwriter := client.BulkWriter(ctx)

	// Go through the different batches of documents
	for {
		// Get a batch of documents
		iter := col.Limit(1).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents adding a delete operation to the BulkWriter
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			_, err = bulkwriter.Delete(doc.Ref)
			if err != nil {
				return err
			}
			numDeleted++
		}

		// If there are no documents to delete the process is over
		if numDeleted == 0 {
			bulkwriter.End()
			break
		}

		bulkwriter.Flush()
	}

	return nil
}

func DeleteHistory() error {
	// Create a Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}

	// Get a BulkWriter that we can add operations to
	col := client.Collection(HISTORY_NAME)
	bulkwriter := client.BulkWriter(ctx)

	for {
		// Get a batch of documents
		iter := col.Limit(1).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents adding a delete operation to the BulkWriter
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			_, err = bulkwriter.Delete(doc.Ref)
			if err != nil {
				return err
			}
			numDeleted++
		}

		// If there are no documents to delete the process is over
		if numDeleted == 0 {
			bulkwriter.End()
			break
		}

		bulkwriter.Flush()
	}

	return nil
}

func DeleteReviews() error {
	// Create a Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}

	// Get a BulkWriter that we can add operations to
	col := client.Collection(REVIEW_NAME)
	bulkwriter := client.BulkWriter(ctx)

	for {
		// Get a batch of documents
		iter := col.Limit(1).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents adding a delete operation to the BulkWriter
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			_, err = bulkwriter.Delete(doc.Ref)
			if err != nil {
				return err
			}
			numDeleted++
		}

		// If there are no documents to delete the process is over
		if numDeleted == 0 {
			bulkwriter.End()
			break
		}

		bulkwriter.Flush()
	}

	return nil
}

func ClearZipStorage() error { // Used to clear the storage of files too big to store in Firebase
	// Create a new client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Get all items
	it := client.Bucket(STORAGE_BUCKET_ID).Objects(ctx, nil)
	// Delete items one by one
	for {
		objAttrs, err := it.Next()
		// No more documents to delete
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// Delete the document
		if err := client.Bucket(STORAGE_BUCKET_ID).Object(objAttrs.Name).Delete(ctx); err != nil {
			log.Printf("Failed to delete object %q: %v", objAttrs.Name, err)
		} else {
			log.Infof("Deleted object %q", objAttrs.Name)
		}
	}

	return nil
}

func CreateReview(userName string, stars int, review string, packageName string) int {
	// Create a Firestore Client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}
	defer client.Close()

	// Create an ID for the review
	reviewID := client.Collection(REVIEW_NAME).NewDoc().ID

	// Query to see if the review already exists
	query := client.Collection(REVIEW_NAME).
		Where("userName", "==", userName).
		Where("packageName", "==", packageName).
		Limit(1)

	// Gets all reviews matching query
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
	}
	// If the user already left a review on the package
	if len(docs) > 0 {
		log.Printf("Review with user name %q and package name %q already exists", userName, packageName)
		return http.StatusConflict
	}

	// Check if the package exists
	query2 := client.Collection(COLLECTION_NAME).
		Where("metadata.Name", "==", packageName).
		Limit(1)

	// Gets all packages matching the query
	docs2, err := query2.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
	}
	// No such package found
	if len(docs2) == 0 {
		log.Printf("No such package")
		return http.StatusNotFound
	}

	// Initialize the Review into a struct
	reviewStruc := map[string]interface{}{
		"userName":    userName,
		"packageName": packageName,
		"review":      review,
		"stars":       stars,
	}

	// Add the new package document to Firestore
	_, err = client.Collection(REVIEW_NAME).Doc(reviewID).Set(ctx, reviewStruc)
	if err != nil {
		log.Critical("Failed to add review to Firestore: %v", err)
		return http.StatusInternalServerError
	}

	// Now add the review action to all of the packages
	query3 := client.Collection(COLLECTION_NAME).
		Where("metadata.Name", "==", packageName)

	// Get all the packages with the name
	docs3, err := query3.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
	}
	// Iterate through each package
	for _, doc := range docs3 {
		data := doc.Data()
		// Get the ID from each package
		if pkgMetadata, ok := data["metadata"].(map[string]interface{}); ok {
			metadata := models.Metadata{Name: pkgMetadata["Name"].(string), ID: pkgMetadata["ID"].(string), Repository: pkgMetadata["Repository"].(string), Version: pkgMetadata["Version"].(string)}
			// Add a REVIEW action
			success := recordActionEntry(client, ctx, "REVIEW", metadata)
			if !success {
				return http.StatusInternalServerError
			}
		}
	}
	return http.StatusCreated
}

func DeleteReview(userName string, packageName string) int {
	// Create a Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}
	defer client.Close()

	// Get the review matching the identifiers
	query := client.Collection(REVIEW_NAME).Where("packageName", "==", packageName).Where("userName", "==", userName)
	docs, _ := query.Documents(ctx).GetAll()
	if len(docs) == 0 {
		log.Println("No reviews found")
		return http.StatusNotFound
	}

	// Delete the review
	batch := client.Batch()
	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}

	// Commit the deletion
	_, err = batch.Commit(ctx)
	if err != nil {
		log.Printf("Failed to commit batch: %v", err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
