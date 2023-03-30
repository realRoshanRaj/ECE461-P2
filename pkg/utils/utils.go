package utils

import "pkgmanager/internal/models"

func extractPackageJsonFromZip(zipfile string) string {
	// extract package.json from zipfile
	return ""
}

func ExtractMetadataFromZip(zipfile string) models.Metadata {
	// pkgJson := extractPackageJsonFromZip(zipfile)

	return models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
}

func ExtractMetadataFromURL(url string) models.Metadata {
	return models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
}
