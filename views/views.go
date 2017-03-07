package views

import (
	"errors"
	"html/template"
	"net/http"
)

// global vars
var (
	tmpl                    map[string]*template.Template // we bundle our templates in a single map of templates
	ErrTemplateDoesNotExist = errors.New("The template does not exist")
)

// load templates on init
// base is our base template calling all other templates
func init() {

	if tmpl == nil {
		tmpl = make(map[string]*template.Template)
	}

	// home page
	tmpl["home"] = template.Must(template.ParseFiles("templates/index.tmpl",
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
	))

	// ebook page
	tmpl["ebook"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/ebook.tmpl",
		"templates/tslisting.tmpl",
	))

	// searchresults page
	tmpl["searchresults"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/recordslist.tmpl",
		"templates/searchresults.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
	))

	// report page after batch sudoc isbn2ppn web service
	tmpl["sudoci2p-report"] = template.Must(template.ParseFiles("templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
		"templates/sudoci2p-report.tmpl",
	))

	// report page after batch sudoc isbn2ppn web service
	tmpl["sudocgetrecords-report"] = template.Must(template.ParseFiles("templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
		"templates/sudocgetrecords-report.tmpl",
	))

	// targetservice page
	tmpl["targetservice"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/package.tmpl",
		"templates/recordslist.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
	))

	// form to create a new target service
	tmpl["targetservicenewget"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
		"templates/tsnew.tmpl",
	))

	// form to update target service
	tmpl["tsupdate"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
		"templates/tsupdate.tmpl",
	))

	// upload page
	tmpl["upload"] = template.Must(template.ParseFiles("templates/upload.tmpl",
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
	))

	// file uploaded
	tmpl["upload-report"] = template.Must(template.ParseFiles("templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
		"templates/upload-report.tmpl",
	))

	// user login form
	tmpl["userlogin"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/userlogin.tmpl",
	))

	// form to create a new user
	tmpl["usernew"] = template.Must(template.ParseFiles(
		"templates/base.tmpl",
		"templates/head.tmpl",
		"templates/nav.tmpl",
		"templates/tslisting.tmpl",
		"templates/tsnew.tmpl",
	))

}

// RenderTmpl is a wrapper around template.ExecuteTemplate
func RenderTmpl(w http.ResponseWriter, name string, data map[string]interface{}) error {

	//make sure template actually exists
	tmpl, ok := tmpl[name]
	if !ok {
		return ErrTemplateDoesNotExist
	}
	tmpl.ExecuteTemplate(w, "base", data)
	return nil
}
