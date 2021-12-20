package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/lizrice/zwiftpower/zp"
	"github.com/spf13/cobra"

	"cloud.google.com/go/storage"
)

var (
	Filename         string
	SpreadsheetID    string
	SpreadsheetSheet string
	Limit            int
	storageClient    *storage.Client
)

const defaultClub = 4516

func getID(args []string, defaultID int) (id int) {
	id = defaultID
	if len(args) >= 1 {
		id64, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Printf("Can't parse ID: %v", err.Error())
			os.Exit(1)
		}
		id = int(id64)
	}
	return id
}

func main() {
	httpCmd := &cobra.Command{
		Use:   "http",
		Short: "Run as a service",
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Unless a filename is specified, assume that this is being written to S3
			if Filename == "" {
				storageClient, err = storage.NewClient(context.Background())
				if err != nil {
					log.Fatalf("storage.NewClient: %v", err)
				}
				log.Printf("Opened storageClient")
			}

			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			http.Handle("/", http.FileServer(http.Dir("/tmp")))
			http.HandleFunc("/trigger", HelloZP)

			// Start HTTP server.
			log.Printf("Listening on port %s", port)
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				log.Fatal(err)
			}
		},
	}

	riderCmd := &cobra.Command{
		Use:   "rider [ID]",
		Short: "Import data for rider ID",
		Run: func(cmd *cobra.Command, args []string) {
			riderID := getID(args, 98588)
			client, err := zp.NewClient()
			if err != nil {
				fmt.Printf("Error getting client: %v", err)
			}

			rider, err := zp.ImportRider(client, riderID)
			if err != nil {
				fmt.Printf("Error getting rider: %v", err)
			}
			fmt.Printf("%v\n", rider.Strings())
		},
	}

	rootCmd := &cobra.Command{
		Use:   "zp [ID]",
		Short: "Import data for club ID",
		Long:  `Default club ID is defaultClub`,
		Run: func(cmd *cobra.Command, args []string) {
			clubID := getID(args, defaultClub)
			err := ZwiftPower(clubID, Limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting ZwiftPower data for %d: %v", clubID, err)
				os.Exit(1)
			}
		},
	}

	os.Setenv("FILENAME", "C:\\Users\\aaron.FULCRUMSW\\test.csv")
	//os.Setenv("SPREADSHEET_ID", "1xoEgu0NwGoNEgQWhFJBlDd9QHPmLk4H0XTGovezVxpI")
	//os.Setenv("SPREADSHEET_SHEET", "Sheet2")
	os.Setenv("LIMIT", "2")

	var limit int
	limitString := os.Getenv("LIMIT")
	if limitString != "" {
		limit, _ = strconv.Atoi(limitString)
	}

	rootCmd.PersistentFlags().StringVarP(&Filename, "filename", "f", os.Getenv("FILENAME"), "Output file name")
	rootCmd.PersistentFlags().StringVarP(&SpreadsheetID, "spreadsheet", "s", os.Getenv("SPREADSHEET_ID"), "Google sheets ID")
	rootCmd.PersistentFlags().StringVarP(&SpreadsheetSheet, "sheetname", "n", os.Getenv("SPREADSHEET_SHEET"), "Google sheets sheet name")
	rootCmd.PersistentFlags().IntVarP(&Limit, "limit", "l", limit, "Restrict to retrieving this number of riders' data. 0 means no limit - get them all.")
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(riderCmd)
	rootCmd.Execute()
}

func setOutput(filename string) (io.WriteCloser, error) {
	ctx := context.Background()

	if SpreadsheetID != "" {
		log.Printf("Writing to spreadsheet")
		sw, err := NewSpreadsheetWriter(ctx, SpreadsheetID, SpreadsheetSheet)
		if err != nil {
			return nil, fmt.Errorf("error getting spreadsheet client: %v", err)
		}
		return sw, nil
	}

	// Upload an object with storage.Writer.
	if storageClient != nil {
		log.Printf("Writing to storage bucket")
		bkt := storageClient.Bucket("revo-rider-aardvark")
		attrs, err := bkt.Attrs(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting bucket attributes: %v", err)
		}

		log.Printf("bucket %s, created at %s, is located in %s with storage class %s\n",
			attrs.Name, attrs.Created, attrs.Location, attrs.StorageClass)
		sc := bkt.Object("results.csv").NewWriter(ctx)
		return sc, nil
	}

	if filename == "" {
		log.Printf("Writing to stdout")
		return os.Stdout, nil
	}

	log.Printf("Writing to file %s", filename)
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file %s: %v\n", filename, err)
	}

	return f, err
}

func ZwiftPower(clubID int, limit int) error {
	client, err := zp.NewClient()
	if err != nil {
		return fmt.Errorf("error getting client: %v", err)
	}

	riders, err := zp.ImportZP(client, clubID)
	if err != nil {
		return fmt.Errorf("error in ImportZP: %v", err)
	}

	f, err := setOutput(Filename)
	if err != nil {
		return fmt.Errorf("opening file %s: %v", Filename, err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("closing: %v", err)
		}
	}()

	writer := NewRowWriter(f)
	defer func() {
		log.Printf("About to flush")
		writer.Flush()
	}()

	for i, rider := range riders {
		var err error
		name := strings.TrimSpace(rider.Name)
		riders[i], err = zp.ImportRider(client, rider.Zwid)
		if err != nil {
			return fmt.Errorf("loading data for %s (%d): %v", name, rider.Zwid, err)
		}
		riders[i].Name = name
		// fmt.Printf("%v\n", riders[i])
		err = writer.WriteRow(riders[i].Strings())
		if err != nil {
			return fmt.Errorf("writing to file: %v", err)
		}

		if limit > 0 && i >= (limit-1) {
			log.Printf("Limiting output to %d riders", limit)
			break
		}
	}

	return nil
}

func HelloZP(w http.ResponseWriter, r *http.Request) {
	clubID := defaultClub
	err := ZwiftPower(clubID, Limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting ZwiftPower data for %d: %v", clubID, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%v", err)))
	}

	fmt.Fprintf(w, "Reading data for %d\n", clubID)
}
