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

func getEntityRef(entity Entity) string {
	if entity.Metadata.Namespace == "default" {
		return fmt.Sprintf("%s:%s", strings.ToLower(entity.Kind), entity.Metadata.Name)
	} else {
		return fmt.Sprintf("%s:%s/%s", strings.ToLower(entity.Kind), entity.Metadata.Namespace, entity.Metadata.Name)
	}
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

func getEntityUrl(entity Entity) string {
	return fmt.Sprintf("%s/catalog/%s/%s/%s", baseUrl, entity.Metadata.Namespace, strings.ToLower(entity.Kind), strings.ToLower(entity.Metadata.Name))
}

func getEntityUrlfromRef(entityRef string) string {
	pattern := `^([^:]+):([^/]+)/([^/]+)$`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(entityRef)
	if matches != nil {
		return fmt.Sprintf("%s/catalog/%s/%s/%s", baseUrl, matches[2], strings.ToLower(matches[1]), strings.ToLower(matches[3]))
	}
	return ""
}

func getKindNamespaceName(entityRef string) (string, string, string) {
	var kind, namespace, name string

	pattern := `^[^:]+:[^/]+/[^/]+$`
	matched, _ := regexp.MatchString(pattern, entityRef)
	if matched {
		ref := strings.Split(entityRef, ":")
		kind = ref[0]
		namespace = strings.Split(ref[1], "/")[0]
		name = strings.Split(ref[1], "/")[1]
	} else {
		fmt.Printf("getKindNamespaceName: %s not a valid entityRef {kind}:{namespace}/{name}", entityRef)
	}

	return kind, namespace, name
}

func parseArgs(args []string) (string, string, string, string) {
	var kind, namespace, name, filter string

	if len(args) > 0 {
		arg := args[0]

		pattern := `^[^:]+:.+$`
		matched, _ := regexp.MatchString(pattern, arg)
		if matched {
			ref := strings.Split(arg, ":")
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
			kind = arg
		}

		allowedKinds := map[string]bool{
			"resource":  true,
			"component": true,
			"system":    true,
			"domain":    true,
			"user":      true,
			"group":     true,
			"location":  true,
		}

		if len(kind) > 0 && kind[len(kind)-1] == 's' {
			kind = kind[:len(kind)-1]
		}
		if !allowedKinds[kind] {
			log.Fatalf("error: backstage doesn't have a resource kind '%s'\nAllowed kinds are: %v", kind, allowedKinds)
		}
	}

	if len(args) > 1 && name == "" {
		name = args[1]
	}

	if kind != "" {
		if len(kind) > 0 && kind[len(kind)-1] == 's' {
			kind = kind[:len(kind)-1]
		}
		filter = fmt.Sprintf("filter=kind=%s", kind)
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

	return kind, namespace, name, filter
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

func displayEntities(header []string, data [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	isNamespaceDefaultOnly := true
	for _, row := range data {
		if row[0] != "default" {
			isNamespaceDefaultOnly = false
			break
		}
	}

	if isNamespaceDefaultOnly {
		fmt.Fprintln(w, strings.Join(header[1:], "\t"))
		for _, row := range data {
			fmt.Fprintln(w, strings.Join(row[1:], "\t"))
		}
	} else {
		fmt.Fprintln(w, strings.Join(header, "\t"))
		for _, row := range data {
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
	}
}
