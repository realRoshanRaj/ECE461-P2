package db

import (
	"context"
	"log"
	"net/http"
	"pkgmanager/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const (
	PROJECT_ID      string = "ece461-project-381318"
	COLLECTION_NAME string = "packages"
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
		return nil, http.StatusConflict
	}

	// Add the new package document to Firestore
	_, err = client.Collection("packages").Doc(documentID).Set(ctx, pkg)
	if err != nil {
		log.Fatalf("Failed to add package to Firestore: %v", err)
		return nil, http.StatusInternalServerError
	}

	return pkg, http.StatusCreated

	// if it does, return error 409 otherwise store package in database
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
