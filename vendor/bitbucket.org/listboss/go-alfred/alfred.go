package Alfred

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path"

    "howett.net/plist"
)

const (
    noResultString string = "No Result Were Found."
    errorTitle     string = "Error in Generating Results."
    settingsFN     string = "settings.plist"
    volatileDir    string = "Library/Caches/com.runningwithcrayons.Alfred-2/Workflow Data"
    dataDir        string = "Library/Application Support/Alfred 2/Workflow Data"
)

type GoAlfred struct {
    bundleID   string
    results    items
    DataDir    string
    BundleDir  string
    CacheDir   string
    SettingsFN string
    id         string
}

type AlfredIcon struct {
    Filename string `xml:",chardata"`
    Type     string `xml:"type,attr,omitempty"`
}

type item struct {
    XMLName      xml.Name   `xml:"item"`
    Uid          string     `xml:"uid,attr,omitempty"`
    Arg          string     `xml:"arg,attr"`
    Type         string     `xml:"type,attr,omitempty"`
    Valid        string     `xml:"valid,attr,omitempty"`
    AutoComplete string     `xml:"autocomplete,attr,omitempty"`
    Title        string     `xml:"title"`
    SubTitle     string     `xml:"subtitle"`
    Icon         AlfredIcon `xml:"icon"`
}

type items struct {
    XMLName xml.Name `xml:"items"`
    Results []item
}

func (ga GoAlfred) Write(p []byte) (n int, err error) {
    return ga.WriteToAlfred()
}

func NewAlfred(id string) *GoAlfred {
    ga := new(GoAlfred)
    ga.init(id)
    return ga
}

func (ga *GoAlfred) init(id string) {
    ga.id = id
    // Get bundleid
    pwd, err := os.Getwd()
    homedir := os.Getenv("HOME")
    if err != nil {
        log.Fatalf("go-alfred: Can't initiate: %v", err)
    }

    ga.BundleDir = pwd
    plistfn := path.Join(ga.BundleDir, "info.plist")
    _, err = os.Stat(plistfn)
    if err != nil {
        log.Printf("Can't locate info.plist: %v\n", plistfn)
    }

    ga.bundleID = ga.getBundleID(plistfn)
    ga.CacheDir = path.Join(homedir, volatileDir, ga.bundleID)
    ga.DataDir = path.Join(homedir, dataDir, ga.bundleID)
    if _, err = os.Stat(ga.CacheDir); err != nil {
        if err = os.MkdirAll(ga.CacheDir, 0755); err != nil {
            log.Fatalf("go-alfred: Can't create cache folder: %v\n", err)
        }
    }
    if _, err = os.Stat(ga.DataDir); err != nil {
        if err = os.MkdirAll(ga.DataDir, 0755); err != nil {
            log.Fatalf("go-alfred: Can't create data folder: %v\n", err)
        }
    }
    ga.SettingsFN = path.Join(ga.DataDir, settingsFN)
}

func (ga *GoAlfred) getBundleID(plistfn string) string {
    buf, err := ioutil.ReadFile(plistfn)
    if err != nil {
        log.Fatalf("%v", err)
    }
    var properties map[string]interface{}
    decoder := plist.NewDecoder(bytes.NewReader(buf))
    err = decoder.Decode(&properties)
    if err != nil {
        log.Fatalf("%v", err)
    }

    v, ok := properties["bundleid"]
    if !ok {
        log.Fatalf("%v doen't contain a 'bundleid' key.", plistfn)
    }

    return (v.(string))
}

func (ga *GoAlfred) XML() (output []byte, err error) {
    // TODO: cache 'output' so we don't call toXML()/xml.Unmarshal unless
    // ga.results has changed
    output, err = ga.results.toXML()
    if err != nil {
        output = nil
    }
    return output, err
}

func (ga *GoAlfred) WriteToAlfred() (n int, err error) {
    var output []byte
    output, err = ga.XML()
    if err != nil {
        ga.MakeError(err)
        output, err = ga.XML()
        if err != nil {
            st := fmt.Sprintf("Error in generating Alfred output: %v",
                err.Error())
            os.Stdout.Write([]byte(st))
            log.Fatal(st)
        }
    }
    n, err = os.Stdout.Write(output)
    // fmt.Println(string(output))
    return n, err
}

func (ga *GoAlfred) MakeError(err error) {
    ga.results = items{}
    subtitle := err.Error()
    ga.AddItem("", errorTitle, subtitle, "no", "no", "", "",
        AlfredIcon{Filename: "erroricon.png"}, false)
}

func (ga *GoAlfred) AddItem(uid, title, subtitle, valid, auto, rtype,
    arg string, icon AlfredIcon, check_valid bool) {

    if title == "" {
        title = noResultString
    }
    r := item{Uid: uid, Arg: arg, Type: rtype, Valid: valid,
        AutoComplete: auto, Title: title, SubTitle: subtitle}
    if check_valid {
        // Make sure the item will work in Alfred as autocomplete if
        // 'auto' parameter is set
        r.make_valid()
    }
    r.Icon = icon
    ga.results.Results = append(ga.results.Results, r)
}

func (results *items) toXML() (output []byte, err error) {
    output, err = xml.MarshalIndent(results, "", "  ")
    if err != nil {
        output = nil
    }
    s := string(output)
    s = `<?xml version="1.0"?>` + s
    return []byte(`<?xml version="1.0"?>` + string(output)), err
}

func (i *item) make_valid() {
    if (i.Valid == "" || i.Valid == "yes") && i.AutoComplete != "" {
        i.Valid = "no"
        i.Arg = ""
    }
}

func NewIcon(fn, itype string) (ico AlfredIcon) {
    return AlfredIcon{Filename: fn, Type: itype}
}
