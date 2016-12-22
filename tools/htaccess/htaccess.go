//
// H T A C E S S - Maintain a .htaccess file
//
// Copyright (C) Philip Schlump, 2012-2015.
// Version: 0.5.9
// BuildNo: 1811
//

//
// Tool to maintain a .htaccess file
//
// How to use
//
// To add a user:
// 		$ htaccess -a username -p password -r realm
//
// To Delete a user:
// 		$ htaccess -d username
//
// To modify a users password:
// 		$ htaccess -m username -p password -r realm
//
// File format is:
// 		user:realm:MD5(user:realm:pass)
// File allows for "#.*" as comments with '#' in column 1
//
// Default file name is .htaccess or can be specified with the -f <name> option on command line.
//
//

package main

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//var opts struct {
//	fileName  string `short:"f" long:"filename"     description:"Path to .htaccess file"      default:".htaccess"`
//	optionAdd string `short:"a" long:"add"          description:"To add user"                 default:""`
//	optionDel string `short:"d" long:"delete"       description:"To delete user"              default:""`
//	opitonMod string `short:"m" long:"modify"       description:"To modify user"              default:""`
//	realm     string `short:"r" long:"realm"         description:"realm name"                  default:""`
//	password  string `short:"p" long:"password"     description:"password"                    default:""`
//}
var fileName = flag.String("filename", ".htaccess", "Path to .htaccess file") // 0
var optionAdd = flag.String("add", "", "To add user")                         // 1	-- Add a new user to the .htaccess file
var optionDel = flag.String("delete", "", "To delete user")                   // 2
var opitonMod = flag.String("modify", "", "To modify user")                   // 3
var realm = flag.String("realm", "", "Realm name")                            // 4
var password = flag.String("password", "", "password")                        // 5
func init() {
	flag.StringVar(fileName, "f", ".htaccess", "Path to .htaccess file") // 0
	flag.StringVar(optionAdd, "a", "", "To add user")                    // 1
	flag.StringVar(optionDel, "d", "", "To delete user")                 // 2
	flag.StringVar(opitonMod, "m", "", "To modify user")                 // 3
	flag.StringVar(realm, "r", "", "Realm name")                         // 4
	flag.StringVar(password, "p", "", "password")                        // 5
}

// ===================================================================================================================================================
func main() {

	flag.Parse()

	// If file not exits - then create
	if !Exists(*fileName) {
		ioutil.WriteFile(*fileName, []byte(""), 0600)
	}

	fr := LoadFile(*fileName)

	if *optionAdd != "" {
		if *realm == "" || *password == "" {
			usage()
		}
		for _, vv := range fr {
			if vv.Un == *optionAdd && vv.Realm == *realm {
				fmt.Printf("Error: Attempt to add when %s already exists in file\n", *optionAdd)
				os.Exit(2)
			}
		}
		fr = append(fr, AddNewEntry(*optionAdd, *realm, *password))
		SaveFile(*fileName, fr)
	} else if *optionDel != "" {
		if *realm == "" {
			usage()
		}
		found := false
		for ii, vv := range fr {
			if vv.Un == *optionDel && vv.Realm == *realm {
				vv.Delete = true
				fr[ii] = vv
				found = true
			}
		}
		if !found {
			fmt.Printf("Error: Attempt to delete non-existend user %s\n", *optionDel)
			os.Exit(2)
		}
		SaveFile(*fileName, fr)
	} else if *opitonMod != "" {
		if *realm == "" || *password == "" {
			usage()
		}
		found := false
		for ii, vv := range fr {
			if vv.Un == *opitonMod && vv.Realm == *realm {
				vv = AddNewEntry(*opitonMod, *realm, *password)
				fr[ii] = vv
				found = true
			}
		}
		if !found {
			fmt.Printf("Error: Attempt to update non-existend user %s\n", *opitonMod)
			os.Exit(2)
		}
		SaveFile(*fileName, fr)
	} else {
		fmt.Printf("Error: Invalid combination of options\n")
		usage()
	}

}

func usage() {
	fmt.Printf(`Usage: htaccess -a user -p passowrd -r realm
     htaccess -d user -r realm
     htaccess -m user -p password r realm
`)
	os.Exit(2)
}

// AnEntry is used to store the entries from the .htaccess file in meory while we work on them
type AnEntry struct {
	IsComment bool
	Delete    bool
	Line      string
	Un        string
	Realm     string
	Pw        string
	LineNo    int
}

// LoadFile will read in the .htaccess file specified by 'fn'.
func LoadFile(fn string) (rv []AnEntry) {
	file, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Errorf("Error opening %q for read, %v", fn, err)
		os.Exit(1)
	}

	rv = make([]AnEntry, 0, 100)
	lines := strings.Split(string(file), "\n")

	for ii, vv := range lines {
		pp := strings.Split(vv, ":")
		ic := false
		if len(vv) == 0 || vv[0] == '#' {
			ic = true
		} else {
			if len(pp) != 3 {
				fmt.Printf("Error: Invalid number of fields on line %d, Line:%s\n", ii+1, vv)
			} else {
				rv = append(rv, AnEntry{IsComment: ic, Line: vv, Un: pp[0], Realm: pp[1], Pw: pp[2], LineNo: ii + 1})
			}
		}
	}
	return
}

// AddNewEntry will add a new user to the in memory AnEntry table in memory
func AddNewEntry(un string, realm string, pw string) (rv AnEntry) {
	rv = AnEntry{
		Un:    un,
		Realm: realm,
		Pw:    md5sum(un + ":" + realm + ":" + pw),
	}
	// rv.Line = rv.Un + ":" + rv.Realm + ":" + rv.Pw
	return
}

// SaveFile will write out the current in memory auth info
func SaveFile(fn string, ff []AnEntry) {
	fp, err := Fopen(fn, "w")
	if err != nil {
		fmt.Printf("Error: Unable to write to %s, error:%s\n", fn, err)
		os.Exit(4)
	}
	defer fp.Close()
	for _, vv := range ff {
		if vv.Delete {
		} else if vv.IsComment {
			fmt.Fprintf(fp, "%s\n", vv.Line)
		} else {
			fmt.Fprintf(fp, "%s:%s:%s\n", vv.Un, vv.Realm, vv.Pw)
		}
	}
}

// Exists return true if the file, specified by 'name', exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Calculate the md5sum of passed string 's'
func md5sum(s string) (buf string) {
	buf = fmt.Sprintf("%x", md5.Sum([]byte(s)))
	return
}

var errInvalidMode = errors.New("Invalid Mode")

// Fopen opens a file with a syntax similar to C fopen
func Fopen(fn string, mode string) (file *os.File, err error) {
	file = nil
	if mode == "r" {
		file, err = os.Open(fn) // For read access.
	} else if mode == "w" {
		file, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	} else if mode == "a" {
		file, err = os.OpenFile(fn, os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			file, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		}
	} else {
		err = errInvalidMode
	}
	return
}

/* vim: set noai ts=4 sw=4: */
