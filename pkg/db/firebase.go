package db

import (
	"context"
	"log"
	"net/http"
	"pkgmanager/internal/models"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PROJECT_ID      string = "ece461-project-381318"
	COLLECTION_NAME string = "packages"
	HISTORY_NAME    string = "history"
)

func CreatePackage(pkg *models.PackageInfo) (*models.PackageInfo, int) {
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

	// Add the new package document to Firestore
	_, err = client.Collection("packages").Doc(documentID).Set(ctx, pkg)
	if err != nil {
		log.Fatalf("Failed to add package to Firestore: %v", err)
		return nil, http.StatusInternalServerError
	}

	success := recordActionEntry(client, ctx, "CREATE", pkg.Metadata)
	if !success {
		return nil, http.StatusInternalServerError
	}

	return pkg, http.StatusCreated

}

func GetPackageByID(id string) (*models.PackageInfo, int) {
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

	success := recordActionEntry(client, ctx, "DOWNLOAD", pkg.Metadata)
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

func GetAllPackages() ([]models.PackageInfo, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		log.Fatalf("Failed to create FireStore Client: %v", err)
		return nil, err
	}

	defer client.Close()
	var pkgs []models.PackageInfo

	itr := client.Collection(COLLECTION_NAME).Documents(ctx)
	for {
		doc, err := itr.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to get all packages: %v", err)
			return nil, err
		}

		tmp_data := models.PackageData{
			Content:   doc.Data()["data"].(map[string]interface{})["Content"].(string),
			URL:       doc.Data()["data"].(map[string]interface{})["URL"].(string),
			JSProgram: doc.Data()["data"].(map[string]interface{})["JSProgram"].(string),
		}

		tmp_meta := models.Metadata{
			Name:    doc.Data()["metadata"].(map[string]interface{})["Name"].(string),
			Version: doc.Data()["metadata"].(map[string]interface{})["Version"].(string),
			ID:      doc.Data()["metadata"].(map[string]interface{})["ID"].(string),
		}

		pkg := models.PackageInfo{
			Data:     tmp_data,
			Metadata: tmp_meta,
		}
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func recordActionEntry(client *firestore.Client, ctx context.Context, action string, metadata models.Metadata) bool {
	historyCollection := client.Collection(HISTORY_NAME)
	newEntry, _, err := historyCollection.Add(ctx, models.ActionEntry{
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
