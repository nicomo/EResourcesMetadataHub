// Package models stores the structs for the objects we have
package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

type Ebook struct {
	DateCreated          time.Time
	DateUpdated          time.Time
	Active               bool
	SfxId                string
	SFXLastHarvest       time.Time
	PublisherLastHarvest time.Time
	SudocLastHarvest     time.Time
	Authors              []string
	Title                string
	Publisher            string
	Pubdate              string
	Edition              int
	Lang                 string
	TargetService        string
	OpenURL              string
	PackageURL           string
	Acquired             bool
	Isbns                []Isbn
	Ppns                 []PPN
	MarcRecord           []string
	Deleted              bool
}

type Isbn struct {
	Isbn       string
	Electronic bool
	Primary    bool
}

type PPN struct {
	Ppn        string
	Electronic bool
	Primary    bool
}

// EbookCreate saves a single ebook to mongo DB
func EbookCreate(ebk Ebook) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getEbooksCol()

	// let's add the time and save
	ebk.DateCreated = time.Now()
	err := coll.Insert(ebk)
	if err != nil {
		logger.Error.Printf("could not save ebook with isbn %v in DB: %s", ebk.Isbns[0], err)
		return err
	}

	return nil
}

// EbookGetByIsbn retrieves an ebook
func EbookGetByIsbns(isbns []string) (Ebook, error) {
	ebk := Ebook{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksCol()

	// construct query
	qry := []string{"{$or: ["}
	qryconditions := make([]string, 0)
	for _, v := range isbns {
		qryconditions = append(qryconditions, "{isbns: {$elemMatch: {isbn: \"", v, "\"}}}")
	}
	qry = append(qry, strings.Join(qryconditions, ","))
	qry = append(qry, "]}")

	logger.Debug.Println(strings.Join(qry, ""))

	// execute query
	err := coll.Find(strings.Join(qry, "")).One(&ebk)
	if err != nil {
		return ebk, err
	}

	return ebk, nil
}

// EbooksCreateOrUpdate checks if ebook exists in DB, using ISBN, then routes to either create or update
func EbooksCreateOrUpdate(records []Ebook) (int, int, error) {

	var createdCounter, updatedCounter int

	for _, record := range records { // for each record

		// pull out the isbns
		isbnsToQuery := make([]string, 0)
		for _, isbn := range record.Isbns { // for each isbn
			if isbn.Isbn != "" {
				isbnsToQuery = append(isbnsToQuery, isbn.Isbn)
			}
		}

		// test if we already know this ebook
		workingRecord, err := EbookGetByIsbns(isbnsToQuery)
		if err != nil { // we don't: none of the isbns were found in DB
			// let's create a new record
			ebkCreateErr := EbookCreate(record)
			if ebkCreateErr != nil {
				logger.Error.Println(ebkCreateErr)
				return createdCounter, updatedCounter, ebkCreateErr
			}
			createdCounter++
		}

		// TODO: we've found the record, let's update.it
		fmt.Println("workingRecord", workingRecord)
	}
	return createdCounter, updatedCounter, nil
}

//TODO: EbookExists returns bool. See https://godoc.org/gopkg.in/mgo.v2#Query.Count

//TODO: EbookUpdate
func EbookUpdate(ebk Ebook) (Ebook, error) {
	return ebk, nil
}

//TODO: EbookSoftDelete
func EbookSoftDelete(ebkId int) error {
	return nil
}

//TODO: EbookDelete
func EbookDelete(ebkId int) error {
	return nil
}
