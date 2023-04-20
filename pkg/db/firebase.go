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

	"github.com/apsystole/log"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	//import semver
)

const (
	PROJECT_ID        string = "ece461-project-381318"
	STORAGE_BUCKET_ID string = "ece461-project-381318.appspot.com"
	COLLECTION_NAME   string = "packages"
	HISTORY_NAME      string = "history"
)

// GetPackageByNameAndVersion returns a package with the given name and version
// package Type is 1 if it is a zip file, 0 if it is a url
func CreatePackage(pkg *models.PackageInfo, contentTooBig bool) (*models.PackageInfo, int) {
	tempContent := pkg.Data.Content

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}

	defer client.Close()

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
		return nil, http.StatusConflict // if package exist, return error 409 otherwise store package in database
	}

	// If the content is too big, store nothing for content
	if contentTooBig {
		// Decode the base64-encoded zip file into a byte array.
		zipData, err := base64.StdEncoding.DecodeString(pkg.Data.Content)
		if err != nil {
			log.Printf("Failed to decode base64 data: %v", err)
			return nil, http.StatusBadRequest
		}

		// log.Println(zipData)

		// Create a new bucket in Firebase Storage to store the zip file.
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create Storage Client: %v", err)
			return nil, http.StatusInternalServerError
		}

		defer storageClient.Close()

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
		// fmt.Println(err)
		log.Critical("Failed to add package to Firestore: %v", err)
		return nil, http.StatusInternalServerError
	}

	success := recordActionEntry(client, ctx, "CREATE", pkg.Metadata)
	if !success {
		return nil, http.StatusInternalServerError
	}

	pkg.Data.Content = tempContent

	return pkg, http.StatusCreated
}

// reason is 1 if it is for download, 0 for rate
func GetPackageByID(id string, reason int) (*models.PackageInfo, int) {
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
	} else if pkg.Data.ContentStorage {
		// Download the zip file from Firebase Storage
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create Storage Client: %v", err)
			return nil, http.StatusInternalServerError
		}

		defer storageClient.Close()

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

		pkg.Data.Content = base64.StdEncoding.EncodeToString(zipData)
	}
	success := recordActionEntry(client, ctx, method, pkg.Metadata)
	if !success {
		return nil, http.StatusInternalServerError
	}

	return &pkg, http.StatusOK
}

func DeletePackageByID(id string) int {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}

	defer client.Close()
	docRef := client.Collection(COLLECTION_NAME).Doc(id)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if !docSnap.Exists() || status.Code(err) == codes.NotFound {
			log.Printf("package with document ID %s not found to delete", id)
			return http.StatusNotFound
		}

		return http.StatusInternalServerError
	}

	_, err = docRef.Delete(ctx)
	if err != nil {
		log.Printf("Failed to delete package with document ID %s", id)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func UpdatePackageByID(id string, newPkg models.PackageInfo) int {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		return http.StatusInternalServerError
	}

	defer client.Close()
	docRef := client.Collection(COLLECTION_NAME).Doc(id)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if !docSnap.Exists() || status.Code(err) == codes.NotFound {
			log.Printf("package with document ID %s not found to delete", id)
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
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		// return http.StatusInternalServerError
	}

	defer client.Close()
	query := client.Collection(HISTORY_NAME).Where("PackageMetadata.Name", "==", package_name)

	var actionEntries []models.ActionEntry
	docs, err := query.Documents(ctx).GetAll()

	if len(docs) == 0 {
		log.Println("No documents found")
		return nil, http.StatusNotFound
	}

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

func DeletePackageByName(package_name string) int {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Printf("Failed to create FireStore Client: %v", err)
		// return http.StatusInternalServerError
	}

	defer client.Close()
	query := client.Collection(HISTORY_NAME).Where("PackageMetadata.Name", "==", package_name)

	docs, err := query.Documents(ctx).GetAll()

	if len(docs) == 0 {
		log.Println("No documents found")
		return http.StatusNotFound
	}

	batch := client.Batch()
	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}

	_, err = batch.Commit(ctx)
	if err != nil {
		log.Printf("Failed to delete documents: %v", err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func GetPackageByRegex(regex string) ([]models.PackageQuery, int) {
	packages, statusCode := GetAllPackages()
	var pkgs []models.PackageQuery
	for _, pkg := range packages {
		matched, err := regexp.MatchString(regex, pkg.Metadata.Name)
		if err != nil {
			return nil, http.StatusInternalServerError
		}

		if matched {
			tmp := models.PackageQuery{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version}
			pkgs = append(pkgs, tmp)
		} else {
			var readme string
			if pkg.Data.URL == "" {
				readme, statusCode = utils.GetReadmeFromZip(pkg.Data.Content)
				if statusCode != http.StatusOK {
					return nil, statusCode
				}
			} else {
				readme, statusCode = utils.GetReadmeTextFromGitHubURL(pkg.Data.URL)
				if statusCode != http.StatusOK {
					return nil, statusCode
				}
			}
			matched, err := regexp.MatchString(regex, readme)
			if err != nil {
				return nil, http.StatusInternalServerError
			}
			if matched {
				tmp := models.PackageQuery{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version}
				pkgs = append(pkgs, tmp)
			}
		}
	}

	if len(pkgs) != 0 {
		return pkgs, statusCode
	} else {
		return nil, http.StatusNotFound
	}
}

func GetAllPackages() ([]models.PackageInfo, int) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Fatalf("Failed to create FireStore Client: %v", err)
		return nil, http.StatusInternalServerError
	}

	defer client.Close()
	var pkgs []models.PackageInfo

	itr := client.Collection(COLLECTION_NAME).Documents(ctx)
	defer itr.Stop()
	for {
		doc, err := itr.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to get all packages: %v", err)
			return nil, http.StatusInternalServerError
		}

		var pkg models.PackageInfo
		if err := doc.DataTo(&pkg); err != nil {
			log.Fatalf("Failed to convert document data: %v", err)
		}
		pkgs = append(pkgs, pkg)
	}

	return pkgs, http.StatusOK
}

func GetPackages(version string, name string, mode string) ([]models.Metadata, int) {
	packages, statusCode := GetAllPackages()
	var pkgs []models.Metadata

	for _, pkg := range packages {
		if mode == "Exact" {
			if pkg.Metadata.Version == version && pkg.Metadata.Name == name {
				// tmp := models.Metadata{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version, ID: pkg.Metadata.ID}
				pkgs = append(pkgs, pkg.Metadata)
			}
		} else if mode == "Bounded range" {
			parts := strings.Split(version, "-")
			lower := parts[0]
			upper := parts[1]
			lowerVersion, _ := semver.NewVersion(lower)
			upperVersion, _ := semver.NewVersion(upper)
			pkgVersion, _ := semver.NewVersion(pkg.Metadata.Version)

			if pkg.Metadata.Name == name && pkgVersion.GreaterThan(lowerVersion) && pkgVersion.LessThan(upperVersion) {
				// tmp := models.Metadata{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version, ID: pkg.Metadata.ID}
				pkgs = append(pkgs, pkg.Metadata)
			}
		} else if mode == "Carat" || mode == "Tilde" {
			carat, _ := semver.NewConstraint(version)
			pkgVersion, _ := semver.NewVersion(pkg.Metadata.Version)

			if pkg.Metadata.Name == name && carat.Check(pkgVersion) {
				// tmp := models.Metadata{Name: pkg.Metadata.Name, Version: pkg.Metadata.Version, ID: pkg.Metadata.ID}
				pkgs = append(pkgs, pkg.Metadata)
			}
		}
	}

	return pkgs, statusCode
}

func recordActionEntry(client *firestore.Client, ctx context.Context, action string, metadata models.Metadata) bool {
	historyCollection := client.Collection(HISTORY_NAME)
	defaultUser := make(map[string]interface{})
	json.Unmarshal([]byte("{\"name\": \"default user\", \"isAdmin\": false}"), &defaultUser)
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

	// Instantiate a client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}

	col := client.Collection(COLLECTION_NAME)
	bulkwriter := client.BulkWriter(ctx)

	for {
		// Get a batch of documents
		iter := col.Limit(1).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to the BulkWriter.
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			bulkwriter.Delete(doc.Ref)
			numDeleted++
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			bulkwriter.End()
			break
		}

		bulkwriter.Flush()
	}

	return nil
}

func DeleteHistory() error {

	// Instantiate a client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}

	col := client.Collection(HISTORY_NAME)
	bulkwriter := client.BulkWriter(ctx)

	for {
		// Get a batch of documents
		iter := col.Limit(1).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to the BulkWriter.
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			bulkwriter.Delete(doc.Ref)
			numDeleted++
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			bulkwriter.End()
			break
		}

		bulkwriter.Flush()
	}

	return nil
}
