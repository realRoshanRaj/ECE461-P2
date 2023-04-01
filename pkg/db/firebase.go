package db

import (
	"context"
	"log"
	"pkgmanager/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type PkgDb interface {
	CreatePackage(pkg *models.PackageInfo) (*models.PackageInfo, error)
	GetAllPackages() ([]models.PackageInfo, error)
}

type Db struct{}

const (
	projectID      string = "ece461-project-381318"
	collectionName string = "packages"
)

func NewPkgDb() PkgDb {
	return &Db{}
}

func (*Db) CreatePackage(pkg *models.PackageInfo) (*models.PackageInfo, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create FireStore Client: %v", err)
		return nil, err
	}

	defer client.Close()
	_, _, err = client.Collection(collectionName).Add(ctx, map[string]interface{}{
		"data":     pkg.Data,
		"metadata": pkg.Metadata,
	})

	if err != nil {
		log.Fatalf("Failed to add a new package: %v", err)
		return nil, err
	}
	return pkg, nil
	// TODO:
	// check if package already exists in database
	// if it does, return error 409 otherwise store package in database
}

func (*Db) GetAllPackages() ([]models.PackageInfo, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create FireStore Client: %v", err)
		return nil, err
	}

	defer client.Close()
	var pkgs []models.PackageInfo

	itr := client.Collection(collectionName).Documents(ctx)
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
