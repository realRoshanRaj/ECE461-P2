package db

import "pkgmanager/internal/models"

func CreatePackageDB(pkg models.PackageInfo) {
	// check if package already exists in database
	// if it does, return error 409
}
