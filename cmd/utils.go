package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

func printYaml(obj interface{}) {
	marshaledYAML, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Println("error marshalling to YAML:", err)
		return
	}
	fmt.Print(string(marshaledYAML))
}

func getRefFromEntity(entity Entity) string {
	if entity.Metadata.Namespace == "default" {
		return fmt.Sprintf("%s:%s", strings.ToLower(entity.Kind), entity.Metadata.Name)
	} else {
		return fmt.Sprintf("%s:%s/%s", strings.ToLower(entity.Kind), entity.Metadata.Namespace, entity.Metadata.Name)
	}
}

func getUrlFromEntity(entity Entity) string {
	return fmt.Sprintf("%s/catalog/%s/%s/%s", baseUrl, entity.Metadata.Namespace, strings.ToLower(entity.Kind), strings.ToLower(entity.Metadata.Name))
}

func getUrlFromRef(entityRef string) string {
	pattern := `^([^:]+):([^/]+)/([^/]+)$`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(entityRef)
	if matches != nil {
		return fmt.Sprintf("%s/catalog/%s/%s/%s", baseUrl, matches[2], strings.ToLower(matches[1]), strings.ToLower(matches[3]))
	}
	return ""
}

func cleanNamespaceDefault(entityRef string) string {
	pattern := `^([^:]+):default/([^/]+)$`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(entityRef)
	if matches != nil {
		return fmt.Sprintf("%s:%s", matches[1], matches[2])
	} else {
		return entityRef
	}
}

func addNamespaceDefault(entityRef string) string {
	pattern := `^([^:]+):([^/]+)$`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(entityRef)
	if matches != nil {
		return fmt.Sprintf("%s:default/%s", matches[1], matches[2])
	}
	return entityRef
}

func getKindNamespaceName(entityRef string) (string, string, string) {
	var kind, namespace, name string

	pattern := `^[^:]+:.+$`
	matched, _ := regexp.MatchString(pattern, entityRef)
	if matched {
		ref := strings.Split(entityRef, ":")
		kind = ref[0]
		pattern := `^[^/]+/[^/]+$`
		matched, _ := regexp.MatchString(pattern, ref[1])
		if matched {
			namespace = strings.Split(ref[1], "/")[0]
			name = strings.Split(ref[1], "/")[1]
		} else {
			namespace = "default"
			name = ref[1]
		}
	} else {
		log.Fatalf("getKindNamespaceName: %s not a valid entityRef {kind}:{namespace}/{name}", entityRef)
	}

	return kind, namespace, name
}

func parseArgs(args []string) string {
	var kinds []string
	var namespace, name, filter string

	if len(args) > 0 {
		arg := args[0]

		pattern := `^[^:]+:.+$`
		matched, _ := regexp.MatchString(pattern, arg)
		if matched {
			ref := strings.Split(arg, ":")
			kinds = append(kinds, ref[0])
			pattern := `^[^/]+/[^/]+$`
			matched, _ := regexp.MatchString(pattern, ref[1])
			if matched {
				namespace = strings.Split(ref[1], "/")[0]
				name = strings.Split(ref[1], "/")[1]
			} else {
				namespace = "default"
				name = ref[1]
			}
		} else {
			kinds = strings.Split(arg, ",")
		}

		allowedKinds := []string{"*", "resource", "component", "system", "domain", "user", "group", "location"}

		for i := range kinds {
			if len(kinds[i]) > 0 && kinds[i][len(kinds[i])-1] == 's' {
				kinds[i] = kinds[i][:len(kinds[i])-1]
			}
			found := false
			for _, allowed := range allowedKinds {
				if allowed == kinds[i] {
					found = true
					break
				}
			}
			if !found {
				log.Fatalf("Error: backstage doesn't have a resource kind '%s'\nAllowed kinds are: %v", kinds[i], allowedKinds)
			}
		}
	}

	if len(args) > 1 && name == "" {
		name = args[1]
	}

	for i := range kinds {
		if kinds[i] == "*" {
			filter = ""
			break
		} else if kinds[i] != "" {
			if filter != "" {
				filter += fmt.Sprintf(",kind=%s", kinds[i])
			} else {
				filter = fmt.Sprintf("filter=kind=%s", kinds[i])
			}
		}
	}
	if namespace != "" {
		if filter != "" {
			filter += fmt.Sprintf(",metadata.namespace=%s", namespace)
		} else {
			filter = fmt.Sprintf("filter=metadata.namespace=%s", namespace)
		}
	}
	if name != "" {
		if filter != "" {
			filter += fmt.Sprintf(",metadata.name=%s", name)
		} else {
			filter = fmt.Sprintf("filter=metadata.name=%s", name)
		}
	}

	return filter
}

func fetchEntitiesByRefs(payload Payload) []Entity {

	var entities []Entity
	var nextCursor string

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("error marshalling payload to JSON: %v", err)
	}

	for {
		url := fmt.Sprintf("%s/api/catalog/entities/by-refs", baseUrl)
		if nextCursor != "" {
			url += fmt.Sprintf("&cursor=%s", nextCursor)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatalf("error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		addAuthHeader(req)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v\n", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("%s", body)
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("error reading response: %v\n", err)
				return nil
			}

			var entitiesResponse EntitiesResponse
			err = json.Unmarshal(body, &entitiesResponse)
			if err != nil {
				fmt.Println("error unmarshalling JSON:", err)
				return nil
			}

			entities = append(entities, entitiesResponse.Items...)

			nextCursor = entitiesResponse.PageInfo.NextCursor
			if nextCursor == "" {
				break
			}
		}
	}

	return entities
}

func fetchEntitiesByQuery(queryParameters string) []Entity {
	var entities []Entity
	var nextCursor string
	for {
		url := fmt.Sprintf("%s/api/catalog/entities/by-query?%s", baseUrl, queryParameters)
		if nextCursor != "" {
			url += fmt.Sprintf("&cursor=%s", nextCursor)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("error creating request: %v\n", err)
			return nil
		}

		addAuthHeader(req)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v\n", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("%s", body)
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("error reading response: %v\n", err)
				return nil
			}

			var entitiesResponse EntitiesResponse
			err = json.Unmarshal(body, &entitiesResponse)
			if err != nil {
				fmt.Println("error unmarshalling JSON:", err)
				return nil
			}

			entities = append(entities, entitiesResponse.Items...)

			nextCursor = entitiesResponse.PageInfo.NextCursor
			if nextCursor == "" {
				break
			}
		}
	}

	return entities
}

func formatOutput(header []string, data [][]string, outputFormat string) {
	if outputFormat == "json" {
		output := make([]map[string]string, len(data))
		for i, row := range data {
			entry := make(map[string]string)
			for j, col := range header {
				if j < len(row) {
					entry[strings.ToLower(col)] = row[j]
				}
			}
			output[i] = entry
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Printf("error marshalling to JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	isNamespaceDefaultOnly := true
	for _, row := range data {
		if len(row) > 0 && row[0] != "default" {
			isNamespaceDefaultOnly = false
			break
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return strings.Join(data[i], "\t") < strings.Join(data[j], "\t")
	})

	if isNamespaceDefaultOnly && len(data) > 0 {
		fmt.Fprintln(w, strings.Join(header[1:], "\t"))
		for _, row := range data {
			if len(row) > 1 {
				fmt.Fprintln(w, strings.Join(row[1:], "\t"))
			}
		}
	} else {
		fmt.Fprintln(w, strings.Join(header, "\t"))
		for _, row := range data {
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
	}
}
