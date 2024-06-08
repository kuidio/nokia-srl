package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func main() {
	templateDir := "templates" // Directory containing your template files
	outputFile := "artifacts/configmap-gotemplates-srl.yaml"

	files, err := os.ReadDir(templateDir)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	templates := make(map[string]string)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Read the content of each template file
		content, err := os.ReadFile(filepath.Join(templateDir, file.Name()))
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", file.Name(), err)
		}

		// Escape the Go template delimiters
		//escapedContent := strings.ReplaceAll(string(content), "{{", "{{`{{")
		//escapedContent = strings.ReplaceAll(escapedContent, "}}", "}}`}}")

		// Store the escaped content in the map
		//templates[file.Name()] = escapedContent
		templates[file.Name()] = string(content)

		fmt.Println(file, "\n", string(content))
	}

	c := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gotemplates-srl",
			Namespace: "kuid-system",
		},
		Data: templates,
	}
	b, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}

	os.WriteFile(outputFile, b, 0644)
}
