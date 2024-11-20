package cmd

type EntitiesResponse struct {
	Items    []Entity `json:"items"`
	PageInfo struct {
		NextCursor string `json:"nextCursor"`
	} `json:"pageInfo"`
	TotalItems int `json:"totalItems"`
}

type Relation struct {
	Type      string `json:"type"`
	TargetRef string `json:"targetRef"`
}

type Entity struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string                 `json:"name"`
		Namespace   string                 `json:"namespace"`
		Description string                 `json:"description"`
		Annotations map[string]interface{} `json:"annotations"`
		Links       []interface{}          `json:"links"`
		Tags        []string               `json:"tags"`
	} `json:"metadata"`
	Relations []Relation             `json:"relations"`
	Spec      map[string]interface{} `json:"spec"`
}

type Entities []Entity
