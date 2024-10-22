package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export -c=[client_name]",
	Short: "Export the configuration of [client_name] to the current folder.",
	Long: `Export the configuration of [client_name] to the current folder.
The configuration will be saved in a file named 'tokendokey.key' in the current directory.

Example:
  export -c=myclient`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			fmt.Println("Error: client name is required")
			return
		}

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)

		// Create tokendokey.key file
		zipFile, err := os.Create("tokendokey.key")
		if err != nil {
			fmt.Println("Error creating tokendokey.key file:", err)
			return
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		// Function to add files to the zip
		addFilesToZip := func(basePath string) error {
			return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				relPath := filepath.Base(path)
				w, err := zipWriter.Create(relPath)
				if err != nil {
					return err
				}

				src, err := os.Open(path)
				if err != nil {
					return err
				}
				defer src.Close()

				_, err = io.Copy(w, src)
				return err
			})
		}

		// Add config directory files to zip
		err = addFilesToZip(configDir)
		if err != nil {
			fmt.Println("Error adding config directory files to tokendokey.key:", err)
			return
		}

		fmt.Println("Configuration file tokendokey.key exported successfully to the current folder")
	},
}

var ImportCmd = &cobra.Command{
	Use:   "import -c=[client_name]",
	Short: "Import the configuration of [client_name] from the current folder.",
	Long: `Import the configuration of [client_name] from the current folder.
The configuration will be loaded from a file named 'tokendokey.key' in the current directory.

Example:
  import -c=myclient`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			fmt.Println("Error: client name is required")
			return
		}

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)
		os.MkdirAll(configDir, os.ModePerm)

		// Check if tokendokey.key file exists
		zipFilePath := "tokendokey.key"
		if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
			fmt.Println("Error: tokendokey.key file does not exist")
			return
		}

		// Unzip the tokendokey.key file
		zipFile, err := zip.OpenReader(zipFilePath)
		if err != nil {
			fmt.Println("Error opening tokendokey.key file:", err)
			return
		}
		defer zipFile.Close()

		for _, file := range zipFile.File {
			dstPath := filepath.Join(configDir, file.Name)

			// Create directories if necessary
			if file.FileInfo().IsDir() {
				os.MkdirAll(dstPath, os.ModePerm)
				continue
			}

			// Create destination file
			dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				fmt.Println("Error creating destination file:", err)
				return
			}
			defer dstFile.Close()

			// Open source file
			srcFile, err := file.Open()
			if err != nil {
				fmt.Println("Error opening source file:", err)
				return
			}
			defer srcFile.Close()

			// Copy file contents
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				fmt.Println("Error copying file:", err)
				return
			}
		}

		fmt.Println("Configuration tokendokey.key imported successfully from the current folder")
	},
}

func init() {
	ExportCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	ExportCmd.MarkFlagRequired("client")

	ImportCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	ImportCmd.MarkFlagRequired("client")
}
