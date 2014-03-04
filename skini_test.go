package skini

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"
    "html/template"
)

// Struct to map input onto
type Config struct {
	Id         string
	LogDir     string
	LogFile    template.HTML
	Supporting []template.HTML

	ServerHttp struct {
		Port   string
		Mode   string
		Keys   []template.HTML
		Colors []string
	}

	Texts     map[string]template.HTML
	Redirects map[string]string
	Press     map[string]map[string]template.HTML
}

func (c *Config) String() string {
	return fmt.Sprintf("Config:\n Id = %v\n LogDir = %v\n LogFile = %v\n ServerHttp.Port = %v\n ServerHttp.Mode = %v\n ServerHttp.Keys = %v\n ServerHttp.Colors = %v\n Texts = %v\n Redirects = %v\n Supporting = %v\n Press = %v\n",
		c.Id,
		c.LogDir,
		c.LogFile,
		c.ServerHttp.Port,
		c.ServerHttp.Mode,
		c.ServerHttp.Keys,
		c.ServerHttp.Colors,
		c.Texts,
		c.Redirects,
		c.Supporting,
        c.Press)
}

// Test inputs
var inputA = `
# Comment
id = /home
logDir = /home/a
logFile = my.log

supporting =
    classA
    classB
    classC

# Comment
[server.http]
    # Server specification
    port = 8080
    # Server specification
    mode = debug
    keys =
        keyOne
        keyTwo
        keyThree
    colors =
        red
        green
        blue

[map.texts]
    html = <br>can have tags</br>
    hello = Hey there !
    bye = See ya :)
    random = s=f(x)
    last/mine = MUST BE PRESENT

[map.redirects]
    ^abc/def$ = efg
    ^bed/bye$ += lko
    MUST_APPPEND

[map.Press | ABC]
    logo = smh.png
    url = /a/b/smh
    keywords =
        apples
        oranges
        persimmons

    blurb += Hello, this is short SMH blurb
            Append ABC Second line.
            Append ABC Third line.


[map.Press | XYZ]
    logo = brw.png
    url = /a/b/brw

    blurb += This BRW blurb is about writing blurbs.
            Append XYZ Second line.
            Append XYZ Third line.

    b += Only one line. EOL.

    c = ######
    keywords =
        carrot
        beetroot
    love = true
    blurb3 += Only one line, next must be new map. EOL.

[map.Press | employers-start ]  
date = Aug 15, 2013           
q = How can I get started with RecruitLoop? 
a += <ol>As an employer you can get started in a number of ways. 
<li>   
 <ins>1.</ins> Post a role and engage an independent recruiter.
Post a hiring project, and receive proposals from the expert independent recruiters in our network. You can engage a      recruiter for any part of the hiring process, as much or little as you need. You can post a role here.
</li>  
<li>   
 <ins>2.</ins> Start with a fixed hiring bundle.                                                                      
Need help in a discrete hiring task, with guaranteed pricing?  You can try a fixed hiring bundle, for a specific task     across a range of recruitment activities. See: What are hiring bundles.
</li>  
<li>
 <ins>3.</ins> Setup video interviews of your candidates.                                                             
You can setup recorded video interviews of your own candidates, using our free recruitment platform. See: How do video    interviews work.              
</li>
</ol>  
        
[map.Press | employers-postrole ]
date = Aug 16, 2013           
q = What happens after I post a role?
a = 16a

[map.Press | ZZZ]
    head = ZZZ Head
    blurb = ZZZ Text
`

var inputB = ""
var inputC = "id = /home/alex"
var inputD = "id = /home/alex\n"
var inputE = "\nid = /home/alex\n"

// Regexes
// [map.name]
var treIsMap = regexp.MustCompile(`^\[\s*map\..*\s*\]$`)
var treMap = regexp.MustCompile(`^\[\s*map\.(?P<name>[a-zA-Z_\.]+)(?P<suffix>\s*\|\s*(?P<key>[a-zA-Z_\.]+))?\s*\]$`)
var treKeyValue = regexp.MustCompile(`^(?P<key>[^=]+)\s+=\s*(?P<value>.*)$`)

// Test regexes
func OFF_TestRegex(t *testing.T) {
	ss := []string{
		"[object.name]",
		"[map]",
		"[map.fruit]",
		"[map.vegies | carrot]",
		"[map.vegies.Grow.BIG | carrot.is.KING ]",
	}
	for j := 0; j < 10; j++ {
		fmt.Println()
	}

	for i, s := range ss {
		fmt.Println()
		fmt.Println(i)
		fmt.Println("---------------------")
		fmt.Println(s)
		ok := runRegex(treMap, s)
		fmt.Println("= ", ok)
	}

	ss = []string{
		"apple = green",
		"kiwis.nz = awesome = buy",
		"^ag23*&@$ = (&^@ = 298",
	}
	for i, s := range ss {
		fmt.Println()
		fmt.Println(i)
		fmt.Println("---------------------")
		fmt.Println(s)
		ok := runRegex(treKeyValue, s)
		fmt.Println("= ", ok)
	}

}

func runRegex(re *regexp.Regexp, line string) bool {
	match := re.FindStringSubmatch(line)
	if match == nil {
		return false
	}

	for i, name := range re.SubexpNames() {
		switch name {
		default:
			fmt.Println("\t", name, "=", match[i])
		}
	}
	return true
}

// Test parsing into structure
//
func TestParseReceiver(t *testing.T) {

	buf := bytes.NewBufferString(inputA)
	if err := parseReceiver(buf); err != nil {
		t.Errorf("Error while parsing: %s", err)
	}

	return

	buf = bytes.NewBufferString(inputB)
	if err := parseReceiver(buf); err != nil {
		//t.Errorf("Error while parsing: %s", err)
	}

	buf = bytes.NewBufferString(inputC)
	if err := parseReceiver(buf); err != nil {
		//t.Errorf("Error while parsing: %s", err)
	}

	buf = bytes.NewBufferString(inputD)
	if err := parseReceiver(buf); err != nil {
		//t.Errorf("Error while parsing: %s", err)
	}

	buf = bytes.NewBufferString(inputE)
	if err := parseReceiver(buf); err != nil {
		//t.Errorf("Error while parsing: %s", err)
	}
}

func parseReceiver(input io.Reader) error {
	fmt.Println("\n\n----------- INPUT FROM STRING -----------")
	fmt.Println(input)
	cfg := Config{}
	if err := Parse(&cfg, input); err != nil {
		fmt.Printf("Error while parsing: %s\n", err)
		return err
	}
	fmt.Println("\n\n----------- OUTPUT FROM STRING -----------")
	fmt.Println(cfg.String())
	return nil
}

// Test parse single file
//
func OFF_TestParseFile(t *testing.T) {
	cfg := Config{}
	if err := ParseFile(&cfg, "test_config_a.ini"); err != nil {
		fmt.Printf("Error while parsing: %s\n", err)
	}
	fmt.Println("\n\n----------- OUTPUT FROM FILE -----------")
	fmt.Println(cfg.String())
}

// Test parsing directory
// Func scans directory and finds a matching INI file
func OFF_TestParseDir(t *testing.T) {
	cfg := Config{}
	err := ParseDir(
		&cfg,
		".",                 // add existing directory
		"test_config_*.ini", // make sure files exist
		"id",
		func(val string) bool {
			fmt.Println("===============>", val)
			if val == "/home" { // set proper matching criteria
				fmt.Println("===============> PASSED <===============")
				return true
			}
			return false
		})

	fmt.Println("\n\n----------- OUTPUT -----------")
	fmt.Println(cfg.String())

	if err != nil {
		t.Errorf("Error while parsing: %s", err)
	}
}
