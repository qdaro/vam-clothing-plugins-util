package lib

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

var managerName = "Stopper.ClothingPluginManager"
var managerPath = managerName + ".latest:/Custom/Scripts/Stopper/ClothingPluginManager/ClothingPluginManager.cs"

func FixVaj(path string, fixOnly bool) *Message {
	uid, err := getUID(path)
	if err != nil {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger", Text: "Couldn't retrieve item's UID.", Details: Ptr(err.Error()),
		}}}
	}

	json, err := os.ReadFile(path)
	if err != nil {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger", Text: "Couldn't read file.", Details: Ptr(err.Error()),
		}}}
	}

	if !gjson.ValidBytes(json) {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger", Text: "Can't parse JSON (invalid).", Details: Ptr(string(json)),
		}}}
	}

	parsed := gjson.ParseBytes(json)
	components := parsed.Get("components")
	storables := parsed.Get("storables")

	if !components.Exists() || !components.IsArray() || !storables.Exists() || !storables.IsArray() {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger",
			Text:    "Invalid JSON.",
			Details: Ptr("\"components\" or \"storables\" properties missing/invalid."),
		}}}
	}

	managerType := "MVRPluginManager"
	isModified := false
	var notes = []Note{}

	// Ensure manager component
	// { "type": "MVRPluginManager" }
	if components.Get(fmt.Sprintf("#(type==\"%s\").type", managerType)).Exists() {
		notes = append(notes, Note{Variant: "info", Text: fmt.Sprintf("%s component already present.", managerType)})
	} else {
		if fixOnly {
			return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
				Variant: "info", Text: "Manager not initialized in this file, skipping.",
			}}}
		}

		managerComponent := map[string]interface{}{"type": managerType}
		json, err = sjson.SetBytes(json, "components.-1", managerComponent)
		if err != nil {
			return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
				Variant: "danger",
				Text:    "Couldn't insert manager component.",
				Details: Ptr(err.Error()),
			}}}
		}

		notes = append(notes, Note{Variant: "success", Text: fmt.Sprintf("Added %s component.", managerType)})
		isModified = true
	}

	// Add/fix storable
	// {
	// 	"id": "mopedlampe:test-glasses",
	// 	"plugins" : {
	// 		"plugin#0" : "Stopper.ClothingPluginManager.7:/Custom/Scripts/Stopper/ClothingPluginManager/ClothingPluginManager.cs"
	// 	}
	// },
	storableIndex := 0
	var storable *gjson.Result

	for i, s := range storables.Array() {
		if IsType[string](s.Get("id").Value()) && IsType[map[string]interface{}](s.Get("plugins").Value()) {
			storableIndex = i
			storable = &s
			break
		}
	}

	if storable == nil {
		data := map[string]interface{}{
			"id":      uid,
			"plugins": map[string]string{"plugin#0": managerPath},
		}

		// OH MY GOD WHY IS WORKING WITH JSON IN GO SO FUCKING PAINFUL!!!
		// If there's more than one storable (100% of the time) we need to do this hacky thing to prepend it.
		res := gjson.GetBytes(json, "storables.0")
		if res.Index > 0 {
			// We insert a "dummy" null element to replace it later, because sjson library
			// can't prepend or insert array items, only replace or append them. And we
			// use sjson because standard and all other libraries randomize/sort map keys...
			json = []byte(string(json)[:res.Index] + "null," + string(json)[res.Index:])

			json, err = sjson.SetBytes(json, "storables.0", data)
			if err != nil {
				return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
					Variant: "danger",
					Text:    "Couldn't insert plugins storable.",
					Details: Ptr(err.Error()),
				}}}
			}
		} else {
			json, err = sjson.SetBytes(json, "storables", []interface{}{data})
			if err != nil {
				return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
					Variant: "danger",
					Text:    "Couldn't insert plugins storable.",
					Details: Ptr(err.Error()),
				}}}
			}
		}

		notes = append(notes, Note{
			Variant: "success",
			Text:    "Added plugins storable.",
			Details: Ptr(JSONMarshalLog(data)),
		})
		isModified = true
	} else {
		// Fix ID & path
		storableId := storable.Get("id").String()
		storableManagerPath := storable.Get("plugins.plugin\\#0").String()
		storablePlugins, isAMap := storable.Get("plugins").Value().(map[string]interface{})

		if storableId == uid && storableManagerPath == managerPath && isAMap && len(storablePlugins) == 1 {
			notes = append(notes, Note{Variant: "info", Text: "Storable ID & manager path are correct."})
		} else {
			old := string(pretty.Pretty([]byte(parsed.Get("storables." + strconv.Itoa(storableIndex)).String())))
			data := map[string]interface{}{
				"id":      uid,
				"plugins": map[string]string{"plugin#0": managerPath},
			}

			json, err = sjson.SetBytes(json, "storables."+strconv.Itoa(storableIndex), data)
			if err != nil {
				return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
					Variant: "danger",
					Text:    "Couldn't replace old storable.",
					Details: Ptr(err.Error()),
				}}}
			}

			new := parsed.Get("storables." + strconv.Itoa(storableIndex)).String()
			notes = append(notes, Note{
				Variant: "success",
				Text:    "Fixed storable ID/path.",
				Details: Ptr(fmt.Sprintf("OLD:\n%v\n\nNEW:\n%s", old, new)),
			})
			isModified = true
		}
	}

	if isModified {
		jsonPretty := pretty.PrettyOptions(json, &pretty.Options{Indent: "\t"})
		err = os.WriteFile(path, jsonPretty, 0644)
		if err != nil {
			notes = append(notes, Note{
				Variant: "danger",
				Text:    "Couldn't write .vaj file.",
				Details: Ptr(err.Error()),
			})
		} else {
			notes = append(notes, Note{
				Variant: "success",
				Text:    "File saved.",
				Details: Ptr(string(pretty.Pretty(json))),
			})
		}
	}

	return &Message{Icon: Ptr("file"), Title: path, Notes: notes}
}

func FixCpl(path string) *Message {
	_, packageName, _, isPrepped := getPreppedPackageName(path)
	if !isPrepped {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "info", Text: "Not in release prep mode, no changes necessary.",
		}}}
	}

	packageNamespace := packageName + ".latest"

	json, err := os.ReadFile(path)
	if err != nil {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger", Text: "Couldn't read file.", Details: Ptr(err.Error()),
		}}}
	}

	newJson := namespaceCustomPaths(json, packageNamespace)
	notes := []Note{}

	if len(newJson) != len(json) {
		notes = append(notes, Note{
			Variant: "success",
			Text:    "Namespaced custom paths to package name.",
			Details: Ptr(fmt.Sprintf("Local \"Custom/*\" and \"SELF:/\" paths in plugin's storables have been namespaced to \"%s:/\".", packageNamespace)),
		})

		jsonPretty := pretty.PrettyOptions(newJson, &pretty.Options{Indent: "\t"})
		err = os.WriteFile(path, jsonPretty, 0644)
		if err != nil {
			notes = append(notes, Note{
				Variant: "danger",
				Text:    "Couldn't write .clothingplugins file.",
				Details: Ptr(err.Error()),
			})
		} else {
			notes = append(notes, Note{
				Variant: "success",
				Text:    "File saved.",
				Details: Ptr(string(pretty.Pretty(newJson))),
			})
		}
	} else {
		notes = append(notes, Note{Variant: "info", Text: "No <code>\"Custom/*\"</code> paths to namespace. All good."})
	}

	return &Message{Icon: Ptr("file"), Title: path, Notes: notes}
}

func FixVap(path string) *Message {
	_, packageName, _, isPrepped := getPreppedPackageName(path)
	if !isPrepped {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "info", Text: "Not in release prep mode, no changes necessary.",
		}}}
	}
	packageNamespace := packageName + ".latest"

	json, err := os.ReadFile(path)
	if err != nil {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger", Text: "Couldn't read file.", Details: Ptr(err.Error()),
		}}}
	}

	notes := []Note{}
	storables := gjson.GetBytes(json, "storables")

	if !storables.Exists() {
		return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
			Variant: "danger", Text: "Invalid .vap file. Missing \"storables\" property.",
		}}}
	}

	managerSuffix := "Stopper.ClothingPluginManager"
	newJson := json

	for i, item := range storables.Array() {
		id := item.Get("id")

		if id.Exists() && strings.HasSuffix(id.String(), managerSuffix) && item.Get("plugins").Exists() {
			prop := fmt.Sprintf("storables.%d", i)
			namespacedJson := namespaceCustomPaths([]byte(item.Raw), packageNamespace)

			newJson, err = sjson.SetRawBytes(newJson, prop, namespacedJson)
			if err != nil {
				return &Message{Icon: Ptr("file"), Title: path, Notes: []Note{{
					Variant: "danger", Text: "Couldn't update storable.", Details: Ptr(err.Error()),
				}}}
			}
		}
	}

	if !bytes.Equal(newJson, json) {
		notes = append(notes, Note{
			Variant: "success",
			Text:    "Namespaced custom paths to package name.",
			Details: Ptr(fmt.Sprintf("Local \"Custom/*\" and \"SELF:/\" paths in plugin's storables have been namespaced to \"%s:/\".", packageNamespace)),
		})

		jsonPretty := pretty.PrettyOptions(newJson, &pretty.Options{Indent: "\t"})
		err = os.WriteFile(path, jsonPretty, 0644)
		if err != nil {
			notes = append(notes, Note{
				Variant: "danger",
				Text:    "Couldn't write .vap file.",
				Details: Ptr(err.Error()),
			})
		} else {
			notes = append(notes, Note{
				Variant: "success",
				Text:    "File saved.",
				Details: Ptr(string(pretty.Pretty(newJson))),
			})
		}
	} else {
		notes = append(notes, Note{Variant: "info", Text: "No <code>\"Custom/*\"</code> paths to namespace. All good."})
	}

	return &Message{Icon: Ptr("file"), Title: path, Notes: notes}
}

var clothingBaseDirExp = regexp.MustCompile(`(?i)(^.*/custom/clothing/(?:female|male)/[^/]+/[^/]+)`)

// Retrieves UID from `.vam` file associated with passed path (can be clothing item's root directory or any file inside it).
// Requires `path` normalized to forward slashes.
// Also requires clothing to use standard `Custom/Clothing/{gender}/{author}/{clothing_name}` folder structure.
func getUID(path string) (string, error) {
	matches := clothingBaseDirExp.FindStringSubmatch(path)
	if matches == nil || len(matches) < 2 {
		return "", fmt.Errorf("getUID: invalid path \"%s\". clothing has to be inside VaM's Custom/Clothing/{gender}/{author}/{name} directory", path)
	}

	dirPath := matches[1]

	dir, err := os.Open(dirPath)
	if err != nil {
		return "", err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return "", err
	}

	vamFile, ok := Find(files, func(file fs.FileInfo) bool {
		return strings.ToLower(filepath.Ext(file.Name())) == ".vam"
	})
	if !ok {
		return "", fmt.Errorf("getUID: couldn't find accompanying .vam file for \"%s\"", path)
	}

	vamFilePath := dirPath + "/" + vamFile.Name()
	data, err := ReadJSON[map[string]interface{}](vamFilePath)
	if err != nil {
		return "", fmt.Errorf("getUID: couldn't read .vam file \"%s\", error: %v", path, err)
	}

	uid, ok := data["uid"].(string)
	if !ok || len(uid) < 3 {
		return "", fmt.Errorf("getUID: couldn't find valid uid property inside .vam file \"%s\"", path)
	}

	return uid, nil
}

var preppedPackageExp = regexp.MustCompile(`(?i)/AddonPackagesBuilder/([^/]+)\.var/.*`)

// Extracts package name from AddonPackagesBuilder path.
func getPreppedPackageName(path string) (full string, authorAndName string, version string, found bool) {
	matches := preppedPackageExp.FindStringSubmatch(path)
	if matches == nil || len(matches) < 2 {
		return "", "", "", false
	}

	parts := strings.Split(matches[1], ".")

	if len(parts) != 3 {
		return "", "", "", false
	}

	return matches[1], strings.Join(parts[:2], "."), parts[2], true
}

var localPathExp = regexp.MustCompile(`(?i)"/?Custom/`)

// Converts all relative paths (`Custom/*`, `SELF:/*`) in a json byte slice to have passed package name as root.
func namespaceCustomPaths(json []byte, packageName string) []byte {
	slash := byte('/')
	name := []byte(packageName + ":/")
	nameLen := len(packageName)
	offset := 0

	for {
		match := FindIndexFromOffset(localPathExp, json, offset)
		if match == nil {
			break
		}

		endIndex := match[0] + 1
		startIndex := endIndex
		if json[startIndex] == slash {
			startIndex++
		}
		json = slices.Concat(json[:endIndex], name, json[startIndex:])
		offset = startIndex + nameLen
	}

	return bytes.ReplaceAll(json, []byte("SELF:/"), []byte(name))
}
