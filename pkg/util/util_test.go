package util

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"reflect"
	"regexp"
	"testing"

	"github.com/redhat-developer/odo/pkg/testingutil"
)

func TestNamespaceOpenShiftObject(t *testing.T) {

	tests := []struct {
		testName        string
		componentName   string
		applicationName string
		want            string
		wantErr         bool
	}{
		{
			testName:        "Test namespacing",
			componentName:   "foo",
			applicationName: "bar",
			want:            "foo-bar",
		},
		{
			testName:        "Blank applicationName with namespacing",
			componentName:   "foo",
			applicationName: "",
			wantErr:         true,
		},
		{
			testName:        "Blank componentName with namespacing",
			componentName:   "",
			applicationName: "bar",
			wantErr:         true,
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := NamespaceOpenShiftObject(tt.componentName, tt.applicationName)

			if tt.wantErr && err == nil {
				t.Errorf("Expected an error, got success")
			} else if tt.wantErr == false && err != nil {
				t.Errorf("Error with namespacing: %s", err)
			}

			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}

}

func TestExtractComponentType(t *testing.T) {

	tests := []struct {
		testName      string
		componentType string
		want          string
		wantErr       bool
	}{
		{
			testName:      "Test namespacing and versioning",
			componentType: "myproject/foo:3.5",
			want:          "foo",
		},
		{
			testName:      "Test versioning",
			componentType: "foo:3.5",
			want:          "foo",
		},
		{
			testName:      "Test plain component type",
			componentType: "foo",
			want:          "foo",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := ExtractComponentType(tt.componentType)
			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}

}

func TestParseComponentImageName(t *testing.T) {

	tests := []struct {
		testName string
		args     string
		want1    string
		want2    string
		want3    string
		want4    string
	}{
		{
			testName: "Case1: Version not specified",
			args:     "nodejs",
			want1:    "nodejs",
			want2:    "nodejs",
			want3:    "nodejs",
			want4:    "latest",
		},
		{
			testName: "Case1: Version not specified",
			args:     "python:3.5",
			want1:    "python:3.5",
			want2:    "python",
			want3:    "python",
			want4:    "3.5",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			got1, got2, got3, got4 := ParseComponentImageName(tt.args)
			if tt.want1 != got1 {
				t.Errorf("Expected imagename to be: %s, got %s", tt.want1, got1)
			}
			if tt.want2 != got2 {
				t.Errorf("Expected component type to be: %s, got %s", tt.want2, got2)
			}
			if tt.want3 != got3 {
				t.Errorf("Expected component name to be: %s, got %s", tt.want3, got3)
			}
			if tt.want4 != got4 {
				t.Errorf("Expected component version to be: %s, got %s", tt.want4, got4)
			}
		})
	}
}

func TestFilePathConversion(t *testing.T) {

	tests := []struct {
		testName  string
		os        string
		direction string
		path      string
		url       string
	}{
		{
			testName:  "Test %q conversion for %q",
			os:        "windows",
			direction: "url to path",
			path:      "c:\\file\\path\\windows\\test",
			url:       "file:///c:/file/path/windows/test",
		},
		{
			testName:  "Test %s conversion for %s",
			os:        "windows",
			direction: "path to url",
			path:      "c:\\file\\path\\windows\\test",
			url:       "file:///c:/file/path/windows/test",
		},
		{
			testName:  "Test %q conversion for %q",
			os:        "linux",
			direction: "url to path",
			path:      "/c/file/path/windows/test",
			url:       "file:///c/file/path/windows/test",
		},
		{
			testName:  "Test %q conversion for %q",
			os:        "linux",
			direction: "path to url",
			path:      "/c/file/path/windows/test",
			url:       "file:///c/file/path/windows/test",
		},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf(tt.testName, tt.direction, tt.os)
		t.Log("Running test: ", testName)
		t.Run(testName, func(t *testing.T) {
			if tt.direction == "url to path" {
				url, err := url.Parse(tt.url)
				if err == nil {
					path := ReadFilePath(url, tt.os)
					if path != tt.path {
						t.Errorf(fmt.Sprintf("Expected an url '%s' to be converted to a path '%s", tt.url, tt.path))
					}
				} else {
					t.Errorf(fmt.Sprintf("Error when parsing url '%s'", tt.url))
				}
			} else if tt.direction == "path to url" {
				url := GenFileURL(tt.path, tt.os)
				if url != tt.url {
					t.Errorf(fmt.Sprintf("Expected a path to be '%s' converted to an url '%s", tt.path, tt.url))
				}
			} else {
				t.Errorf(fmt.Sprintf("Unexpected direction '%s'", tt.direction))
			}
		})
	}
}

func TestParametersAsMap(t *testing.T) {

	tests := []struct {
		testName    string
		sliceInput  []string
		expectedMap map[string]string
	}{
		{
			testName:    "empty slice",
			sliceInput:  []string{},
			expectedMap: map[string]string{},
		},
		{
			testName:   "slice with single element",
			sliceInput: []string{"name=value"},
			expectedMap: map[string]string{
				"name": "value",
			},
		},
		{
			testName:   "slice with multiple elements",
			sliceInput: []string{"name1=value1", "name2=value2", "name3=value3"},
			expectedMap: map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			resultingMap := ConvertKeyValueStringToMap(tt.sliceInput)
			if !reflect.DeepEqual(tt.expectedMap, resultingMap) {
				t.Errorf("Expected %s, got %s", tt.expectedMap, resultingMap)
			}
		})
	}

}

func TestGetDNS1123Name(t *testing.T) {

	tests := []struct {
		testName string
		param    string
		want     string
	}{
		{
			testName: "Case 1: Test get DNS-1123 name for namespace and version qualified imagestream",
			param:    "myproject/foo:3.5",
			want:     "myproject-foo-3-5",
		},
		{
			testName: "Case 2: Test get DNS-1123 name for doubly hyphenated string",
			param:    "nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 3: Test get DNS-1123 name for string starting with underscore",
			param:    "_nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 4: Test get DNS-1123 name for string ending with underscore",
			param:    "nodejs--myproject-foo-3.5_",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 5: Test get DNS-1123 name for string having multiple non alpha-numeric chars as prefix",
			param:    "_#*nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 6: Test get DNS-1123 name for string having multiple non alpha-numeric chars as suffix",
			param:    "nodejs--myproject-foo-3.5=_@",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 7: Test get DNS-1123 name for string having with multiple non alpha-numeric chars as prefix and suffix",
			param:    " _#*nodejs--myproject-foo-3.5=_@ ",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 8: Test get DNS-1123 should remove invalid chars",
			param:    "myproject/$foo@@:3.5",
			want:     "myproject-foo-3-5",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := GetDNS1123Name(tt.param)
			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}

}

func TestGetRandomName(t *testing.T) {
	type args struct {
		prefix    string
		existList []string
	}
	tests := []struct {
		testName string
		args     args
		// want is regexp if expectConflictResolution is true else its a full name
		want string
	}{
		{
			testName: "Case: Optional suffix passed and prefix-suffix as a name is not already used",
			args: args{
				prefix: "odo",
				existList: []string{
					"odo-auth",
					"odo-pqrs",
				},
			},
			want: "odo-[a-z]{4}",
		},
		{
			testName: "Case: Optional suffix passed and prefix-suffix as a name is already used",
			args: args{
				prefix: "nodejs-ex-nodejs",
				existList: []string{
					"nodejs-ex-nodejs-yvrp",
					"nodejs-ex-nodejs-abcd",
				},
			},
			want: "nodejs-ex-nodejs-[a-z]{4}",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := GetRandomName(tt.args.prefix, -1, tt.args.existList, 3)
			if err != nil {
				t.Errorf("failed to generate a random name. Error %v", err)
			}

			r, _ := regexp.Compile(tt.want)
			match := r.MatchString(name)
			if !match {
				t.Errorf("Received name %s which does not match %s", name, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		testName  string
		str       string
		strLength int
		want      string
	}{
		{
			testName:  "Case: Truncate string to greater length",
			str:       "qw",
			strLength: 4,
			want:      "qw",
		},
		{
			testName:  "Case: Truncate string to lesser length",
			str:       "rtyu",
			strLength: 3,
			want:      "rty",
		},
		{
			testName:  "Case: Truncate string to -1 length",
			str:       "Odo",
			strLength: -1,
			want:      "Odo",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			receivedStr := TruncateString(tt.str, tt.strLength)
			if tt.want != receivedStr {
				t.Errorf("Truncated string %s is not same as %s", receivedStr, tt.want)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		testName  string
		strLength int
	}{
		{
			testName:  "Case: Generate random string of length 4",
			strLength: 4,
		},
		{
			testName:  "Case: Generate random string of length 3",
			strLength: 3,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := GenerateRandomString(tt.strLength)
			r, _ := regexp.Compile(fmt.Sprintf("[a-z]{%d}", tt.strLength))
			match := r.MatchString(name)
			if !match {
				t.Errorf("Randomly generated string %s which does not match regexp %s", name, fmt.Sprintf("[a-z]{%d}", tt.strLength))
			}
		})
	}
}

func TestSliceDifference(t *testing.T) {
	tests := []struct {
		testName       string
		slice1         []string
		slice2         []string
		expectedResult []string
	}{
		{
			testName:       "Empty slices",
			slice1:         []string{},
			slice2:         []string{},
			expectedResult: []string{},
		},
		{
			testName:       "Single different slices",
			slice1:         []string{"a"},
			slice2:         []string{"b"},
			expectedResult: []string{"b"},
		},
		{
			testName:       "Single same slices",
			slice1:         []string{"a"},
			slice2:         []string{"a"},
			expectedResult: []string{},
		},
		{
			testName:       "Large slices with matching and non matching items",
			slice1:         []string{"a", "b", "c", "d", "e"},
			slice2:         []string{"e", "a", "u", "1", "d"},
			expectedResult: []string{"u", "1"},
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			result := SliceDifference(tt.slice1, tt.slice2)
			if !reflect.DeepEqual(tt.expectedResult, result) {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGetAbsPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		absPath string
		wantErr bool
	}{
		{
			name:    "Case 1: Valid abs path resolution of `~`",
			path:    "~",
			wantErr: false,
		},
		{
			name:    "Case 2: Valid abs path resolution of `.`",
			path:    ".",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			switch tt.path {
			case "~":
				usr, err := user.Current()
				if err != nil {
					t.Errorf("Failed to get absolute path corresponding to `~`. Error %v", err)
					return
				}
				tt.absPath = usr.HomeDir
			case ".":
				absPath, err := os.Getwd()
				if err != nil {
					t.Errorf("Failed to get absolute path corresponding to `.`. Error %v", err)
					return
				}
				tt.absPath = absPath
			}
			result, err := GetAbsPath(tt.path)
			if result != tt.absPath {
				t.Errorf("Expected %v, got %v", tt.absPath, result)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("Expected error: %v got error %v", tt.wantErr, err)
			}
		})
	}
}

func TestCheckPathExists(t *testing.T) {
	dir, err := ioutil.TempFile("", "")
	defer os.RemoveAll(dir.Name())
	if err != nil {
		return
	}
	tests := []struct {
		fileName string
		wantBool bool
	}{
		{
			fileName: dir.Name(),
			wantBool: true,
		},
		{
			fileName: dir.Name() + "-blah",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		exists := CheckPathExists(tt.fileName)
		if tt.wantBool != exists {
			t.Errorf("the expected value of TestCheckPathExists function is different : %v, got: %v", tt.wantBool, exists)
		}
	}
}

func TestGetHostWithPort(t *testing.T) {

	tests := []struct {
		inputURL string
		want     string
		wantErr  bool
	}{
		{
			inputURL: "https://example.com",
			want:     "example.com:443",
			wantErr:  false,
		},
		{
			inputURL: "https://example.com:8443",
			want:     "example.com:8443",
			wantErr:  false,
		},
		{
			inputURL: "http://example.com",
			want:     "example.com:80",
			wantErr:  false,
		},
		{
			inputURL: "notexisting://example.com",
			want:     "",
			wantErr:  true,
		},
		{
			inputURL: "http://127.0.0.1",
			want:     "127.0.0.1:80",
			wantErr:  false,
		},
		{
			inputURL: "example.com:1234",
			want:     "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Testing inputURL: %s", tt.inputURL), func(t *testing.T) {
			got, err := GetHostWithPort(tt.inputURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostWithPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostWithPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIgnoreRulesFromDirectory(t *testing.T) {
	testDir, err := ioutil.TempDir(os.TempDir(), "odo-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)
	tests := []struct {
		name             string
		directoryName    string
		filesToCreate    []string
		rulesOnGitIgnore string
		rulesOnOdoIgnore string
		wantRules        []string
		wantErr          bool
	}{
		{
			name:             "test case 1: no odoignore and no gitignore",
			directoryName:    testDir,
			filesToCreate:    []string{""},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 2: no odoignore but gitignore exists with no rules",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore"},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 3: no odoignore but gitignore exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore"},
			rulesOnGitIgnore: "*.js\n\n/openshift/**/*.json\n/tests",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git", "*.js", "/openshift/**/*.json", "/tests"},
			wantErr:          false,
		},
		{
			name:             "test case 4: odoignore exists with no rules",
			directoryName:    testDir,
			filesToCreate:    []string{".odoignore"},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 5: odoignore exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{".odoignore"},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "*.json\n\n/openshift/**/*.js",
			wantRules:        []string{".git", "*.json", "/openshift/**/*.js"},
			wantErr:          false,
		},
		{
			name:             "test case 6: odoignore and gitignore both exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore", ".odoignore"},
			rulesOnGitIgnore: "/tests",
			rulesOnOdoIgnore: "*.json\n\n/openshift/**/*.js",
			wantRules:        []string{".git", "*.json", "/openshift/**/*.js"},
			wantErr:          false,
		},
		{
			name:             "test case 7: no odoignore but gitignore exists with rules and comments",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore"},
			rulesOnGitIgnore: "*.js\n\n/openshift/**/*.json\n\n\n#/tests",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git", "*.js", "/openshift/**/*.json"},
			wantErr:          false,
		},
		{
			name:             "test case 8: odoignore exists exists with rules and comments",
			directoryName:    testDir,
			filesToCreate:    []string{".odoignore"},
			rulesOnOdoIgnore: "*.js\n\n\n/openshift/**/*.json\n\n\n#/tests\n/bin",
			rulesOnGitIgnore: "",
			wantRules:        []string{".git", "*.js", "/openshift/**/*.json", "/bin"},
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		for _, fileName := range tt.filesToCreate {
			var err error
			if fileName == ".gitignore" {
				err = testingutil.MakeFileWithContent(testDir, fileName, tt.rulesOnGitIgnore)
			} else if fileName == ".odoignore" {
				err = testingutil.MakeFileWithContent(testDir, fileName, tt.rulesOnOdoIgnore)
			}
			if err != nil {
				t.Fatal(err)
			}
		}

		gotRules, err := GetIgnoreRulesFromDirectory(testDir)

		if err == nil && !tt.wantErr {
			if !reflect.DeepEqual(gotRules, tt.wantRules) {
				t.Errorf("the expected value of rules are different, excepted: %v, got: %v", tt.wantRules, gotRules)
			}
		} else if err == nil && tt.wantErr {
			t.Error("error was expected, but no error was returned")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
		}
		err = testingutil.RemoveContentsFromDir(testDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}
