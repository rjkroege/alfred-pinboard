package Alfred

import (
    // "fmt"
    "testing"
    "time"
)

func TestBasics(t *testing.T) {
    tests := []struct {
        id       string
        expected string
    }{
        {id: "TestBasic", expected: `<?xml version="1.0"?><items></items>`},
    }
    var ga *GoAlfred
    for _, test := range tests {
        ga = NewAlfred(test.id)
        res, err := ga.XML()
        sres := string(res)
        if err != nil {
            t.Fatalf("%s has faild with: %v", test.id, err)
        }
        if sres != test.expected {
            t.Errorf("Expected %v but received %v\n", test.expected, sres)
        }
    }
}

func TestSettings(t *testing.T) {
    ga := NewAlfred("TestSettings")

    err := ga.Set("AlfredApp", "yes")
    if err != nil {
        t.Errorf("Couldn't write to settings file\n%v\n", err)
    }

    r, err := ga.Get("AlfredApp")
    if err != nil {
        t.Errorf("Couldn't read 'AlfredApp' key from settings file\n%v\n", err)
    }
    if r != "yes" {
        ferror(t, "yes", r)
    }

    err = ga.Set("username", "password")
    if err != nil {
        t.Errorf("Couldn't set the second value.\n%v\n", err)
    }
    s, err := ga.Get("username")
    if err != nil {
        t.Errorf("Couldn't read 'username' key from settings file.\n%v\n", err)
    }
    if s != "password" {
        ferror(t, "password", s)
    }

    // set/get date
    time_ := time.Now()
    if err = ga.Set("Time1", time_.Format(time.RFC3339Nano)); err != nil {
        t.Errorf("Couldn't set a time value as string.\n%v\n", err)
    }
    time.Sleep(10 * time.Millisecond)

    tr, err := ga.Get("Time1")
    if err != nil {
        t.Errorf("Couldn't read 'Time1' key from settings file.\n%v\n", err)
    }

    mytime, err := time.Parse(time.RFC3339Nano, tr)
    if err != nil {
        t.Errorf("Error in parsing the time from file.\n%v\n", err)
    }

    if !mytime.Equal(time_) {
        t.Errorf("Read/Set times are not equal:\n%v\n%v\n", mytime, time_)
    }

    // change a setting
    if err = ga.Set("AlfredApp", "changed"); err != nil {
        t.Errorf("Couldn't re-set a key.\n%v\n", err)
    }
    rs, err := ga.Get("AlfredApp")
    if err != nil {
        t.Errorf("Couldn't read 'AlfredApp' key from settings file.\n%v\n", err)
    }
    if rs != "changed" {
        ferror(t, "changed", rs)
    }
}

func TestAddItem(t *testing.T) {
    var ga *GoAlfred
    icon := NewIcon("pin.png", "icontype")
    ga = NewAlfred("TestAddItem")

    var tests = []struct {
        itemargs   []string
        make_valid bool
        expected   string
    }{
        {itemargs: []string{"uiduidadc", "TestBasic Title", "Adding stuff.", "yes", "yes", "file", "deleteme"},
            make_valid: false,
            expected: `<?xml version="1.0"?><items>
  <item uid="uiduidadc" arg="deleteme" type="file" valid="yes" autocomplete="yes">
    <title>TestBasic Title</title>
    <subtitle>Adding stuff.</subtitle>
    <icon type="icontype">pin.png</icon>
  </item>
</items>`,
        },
        {itemargs: []string{"uiduidadc", "TestBasic Title", "Adding stuff.", "yes", "yes", "file", "deleteme"},
            make_valid: true,
            expected: `<?xml version="1.0"?><items>
  <item uid="uiduidadc" arg="deleteme" type="file" valid="yes" autocomplete="yes">
    <title>TestBasic Title</title>
    <subtitle>Adding stuff.</subtitle>
    <icon type="icontype">pin.png</icon>
  </item>
  <item uid="uiduidadc" arg="" type="file" valid="no" autocomplete="yes">
    <title>TestBasic Title</title>
    <subtitle>Adding stuff.</subtitle>
    <icon type="icontype">pin.png</icon>
  </item>
</items>`,
        },
        {itemargs: []string{"", "", "Adding stuff.", "yes", "yes", "file", "deleteme"},
            make_valid: true,
            expected: `<?xml version="1.0"?><items>
  <item uid="uiduidadc" arg="deleteme" type="file" valid="yes" autocomplete="yes">
    <title>TestBasic Title</title>
    <subtitle>Adding stuff.</subtitle>
    <icon type="icontype">pin.png</icon>
  </item>
  <item uid="uiduidadc" arg="" type="file" valid="no" autocomplete="yes">
    <title>TestBasic Title</title>
    <subtitle>Adding stuff.</subtitle>
    <icon type="icontype">pin.png</icon>
  </item>
  <item arg="" type="file" valid="no" autocomplete="yes">
    <title>No Result Were Found.</title>
    <subtitle>Adding stuff.</subtitle>
    <icon type="icontype">pin.png</icon>
  </item>
</items>`,
        },
    }
    for _, test := range tests {
        args := make([]string, 7)
        for i, a := range test.itemargs {
            args[i] = a
        }
        ga.AddItem(args[0], args[1], args[2], args[3], args[4], args[5],
            args[6], icon, test.make_valid)
        res, err := ga.XML()
        if err != nil {
            t.Fatalf("%s has faild with: %v", "TestAddItem", err)
        }
        if string(res) != test.expected {
            ferror(t, test.expected, string(res))
        }
    }
}

func TestMakeError(t *testing.T) {
    ga := NewAlfred("TestMakeError")
    ga.AddItem("uid", "title", "subtitle", "yes", "yes", "file", "arg",
        AlfredIcon{}, false)
    ga.MakeError(AlfredError("Testing Forcing an error result."))
    rec, err := ga.XML()
    if err != nil {
        t.Fatalf("%s has faild with: %v", "TestMakeError", err)
    }
    expected := `<?xml version="1.0"?><items>
  <item arg="" valid="no" autocomplete="no">
    <title>Error in Generating Results.</title>
    <subtitle>Testing Forcing an error result.</subtitle>
    <icon>erroricon.png</icon>
  </item>
</items>`
    if string(rec) != expected {
        t.Errorf("Expected\n%v\nbut received ->\n%v\n", expected, string(rec))
    }
}

func ferror(t *testing.T, exp, rec interface{}) {
    t.Errorf("Expected\n%v\nbut received ->\n%v\n", exp.(string), rec.(string))
    t.Fail()
}
