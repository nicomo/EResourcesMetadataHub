package controllers

import (
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

// createTSStructFromForm creates a TS struct from a form
func createTSStructFromForm(r *http.Request) (models.TargetService, error) {
	// init our Target Service struct
	ts := models.TargetService{}

	// used by gorilla schema to parse html forms
	decoder := schema.NewDecoder()

	// we parse the form
	parseErr := r.ParseForm()
	if parseErr != nil {
		logger.Error.Println(parseErr)
		return ts, parseErr
	}

	// r.PostForm is a map of our POST form values
	// we create a struct from form
	// but ignore the fields which do not exist in the struct
	decoder.IgnoreUnknownKeys(true)
	errDecode := decoder.Decode(&ts, r.PostForm)
	if errDecode != nil {
		logger.Error.Println(errDecode)
		return ts, errDecode
	}

	// parse the csv conf part of the form manually
	csvConf, ok := TargetServiceNewCSVConf(r.Form)
	if !ok {
		logger.Info.Printf("no csv conf created for TS %s", ts.TSName)
	} else {
		ts.TSCsvConf = csvConf
	}

	return ts, nil
}

// csvConfConvert swaps keys and values of the TSCSVConf struct
func csvConfConvert(c models.TSCSVConf) map[int]string {

	sc := make(map[int]string)

	s := reflect.ValueOf(c)
	typeOfc := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		typ := f.Kind().String()
		switch typ {
		case "slice": // authors
			for _, v := range f.Interface().([]int) {
				sc[v] = typeOfc.Field(i).Name
			}
		case "int":
			myi := f.Interface().(int)
			if myi == 0 {
				continue
			}
			sc[myi] = typeOfc.Field(i).Name
		}
	}
	return sc
}

// csvConfGetNFields returns the number of fields used in a particular TSCVConf struct
func csvConfGetNFields(c models.TSCSVConf) int {
	m := csvConfConvert(c)
	return len(m)
}

// csvConf2String returns the csvConf as a string to be displayed in UI
func csvConf2String(c map[int]string) string {

	var csvConfString string

	// To store the keys in slice in sorted order
	var keys []int
	for k := range c {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// To perform the opertion you want
	for _, k := range keys {
		csvConfString += c[k] + "; "
	}

	return csvConfString
}

// TargetServiceHandler retrieves the ebooks linked to a Target Service
//  and various other info, e.g. number of library records linked, etc.
func TargetServiceHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// package name is last part of the URL
	tsname := r.URL.Path[len("/package/"):]
	d["myPackage"] = tsname

	// get the TS Struct from DB
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}
	d["IsTSActive"] = myTS.TSActive

	// if there is one, get the csv conf in a string for display
	convertedCsvConf := csvConfConvert(myTS.TSCsvConf)
	// we do have a csv configuration for this TS
	if len(convertedCsvConf) > 0 {
		d["myPackageCSVConf"] = csvConf2String(convertedCsvConf)
	}

	// any ebooks records have this TS?
	count := models.TSCountEbooks(tsname)
	d["myPackageEbooksCount"] = count

	if count > 0 { // no need to query for actual ebooks otherwise

		// how many ebooks have marc records
		nbRecordsUnimarc := models.TSCountRecordsUnimarc(tsname)
		d["myPackageRecordsUnimarcCount"] = nbRecordsUnimarc

		// how many ebooks have a PPN from the Sudoc Union Catalog
		nbPPNs := models.TSCountPPNs(tsname)
		d["myPackagePPNsCount"] = nbPPNs

		// get the ebooks
		records, err := models.EbooksGetByTSName(tsname)
		if err != nil {
			logger.Error.Println(err)
		}
		d["myRecords"] = records
	}

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "targetservice", d)
}

// TargetServiceUpdateGetHandler fills the update form for a Target Service
func TargetServiceUpdateGetHandler(w http.ResponseWriter, r *http.Request) {

	tsname := r.URL.Path[len("/package/update/"):]

	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// retrieve Target Service Struct
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	d["myTS"] = myTS

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "tsupdate", d)
}

// TargetServiceUpdatePostHandler updates a target service
func TargetServiceUpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	tsname := r.URL.Path[len("/package/update/"):]
	d := make(map[string]interface{})

	ts, formErr := createTSStructFromForm(r)
	if formErr != nil {
		d["tsUpdateErr"] = formErr
		logger.Error.Println(formErr)
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	tsToUpdate, tsToUpdateErr := models.GetTargetService(tsname)
	if tsToUpdateErr != nil {
		logger.Error.Println(tsToUpdateErr)
		d["tsUpdateErr"] = tsToUpdateErr
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	ts.ID = tsToUpdate.ID

	err := models.TSUpdate(ts)
	if err != nil {
		d["tsUpdateErr"] = err
		logger.Error.Println(err)
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	http.Redirect(w, r, "/", 303)
}

// TargetServiceNewCSVConf  has the logic for parsing the new TS form and
// extracting the values to create a new csv configuration struct
func TargetServiceNewCSVConf(form url.Values) (models.TSCSVConf, bool) {
	conf := models.TSCSVConf{}

	nfields := 0
	var authors []int

	for k, v := range form { // url.Values is a map
		for _, w := range v { // and each value is in a []string
			switch {
			// index from 1 to keep 0 as nil value
			// so when used later to read a csv file, use as value-1
			// see csvio.go
			case w == "author":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				authors = append(authors, i)
				nfields++
			case w == "eisbn":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Eisbn = i
				nfields++
			case w == "edition":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Edition = i
				nfields++
			case w == "isbn":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Isbn = i
				nfields++
			case w == "lang":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Lang = i
				nfields++
			case w == "publisher":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Publisher = i
				nfields++
			case w == "pubdate":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Pubdate = i
				nfields++
			case w == "title":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				conf.Title = i
				nfields++
			case w == "url":
				i, err := strconv.Atoi(k)
				if err != nil {
					logger.Error.Println(err)
				}
				nfields++
				conf.URL = i
			default:
				continue
			}
		}
	}

	if len(authors) > 0 {
		conf.Authors = authors
	}

	if nfields == 0 {
		return conf, false
	}

	return conf, true
}

// TargetServiceNewGetHandler displays the form to register a new Target Service (i.e. ebook package)
func TargetServiceNewGetHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing
	views.RenderTmpl(w, "targetservicenewget", d)
}

// TargetServiceNewPostHandler manages the form to register a new Target Service (i.e. ebook package)
func TargetServiceNewPostHandler(w http.ResponseWriter, r *http.Request) {
	d := make(map[string]interface{})

	ts, formErr := createTSStructFromForm(r)
	if formErr != nil {
		d["tsCreateErr"] = formErr
		logger.Error.Println(formErr)
		views.RenderTmpl(w, "targetservicenewget", d)
		return
	}

	err := models.TSCreate(ts)
	if err != nil {
		d["tsCreateErr"] = err
		logger.Error.Println(err)
		views.RenderTmpl(w, "targetservicenewget", d)
		return
	}

	http.Redirect(w, r, "/", 303)
}

// TargetServiceToggleActiveHandler changes the boolean "active" for a TS *and* records who are linked to *only* this TS
func TargetServiceToggleActiveHandler(w http.ResponseWriter, r *http.Request) {

	// package name is last part of the URL
	tsname := r.URL.Path[len("/package/toggleactive/"):]

	// retrieve Target Service Struct
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// retrieve records with thats TS
	records, err := models.EbooksGetByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// change "active" bool in those records
	// and save each to DB
	for _, v := range records {
		if myTS.TSActive {
			v.Active = false
		} else {
			v.Active = true
		}
		_, vUpdateErr := models.EbookUpdate(v)
		if vUpdateErr != nil {
			logger.Error.Printf("can't update record %v: %v", v.ID, vUpdateErr)
		}
	}

	// change "active" bool in TS struct
	if myTS.TSActive {
		myTS.TSActive = false
	} else {
		myTS.TSActive = true
	}

	// save TS to DB
	tsUpdateErr := models.TSUpdate(myTS)
	if tsUpdateErr != nil {
		logger.Error.Println(tsUpdateErr)
	}

	// refresh TS page
	urlStr := "/package/" + tsname
	http.Redirect(w, r, urlStr, 303)
}
