package projectInit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
)

type template struct {
	Name string
}

func (t template) GetName() string {
	return t.Name[:len(t.Name)-len(filepath.Ext(t.Name))]
}

// ListTemplates lists all the Gauge templates available in GaugeTemplatesURL
func ListTemplates() {
	templatesURL := config.GaugeTemplatesUrl()
	_, err := common.UrlExists(templatesURL)
	if err != nil {
		logger.Fatalf("Gauge templates URL is not reachable: %s", err.Error())
	}
	res, err := http.Get(templatesURL)
	if err != nil {
		logger.Fatalf("Error occurred while downloading templates list: %s", err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		logger.Errorf(fmt.Sprintf("Error downloading templates info: %s.\n%s", templatesURL, res.Status))
	}
	templates := []template{}
	json.NewDecoder(res.Body).Decode(&templates)
	for _, t := range templates {
		logger.Info(t.GetName())
	}
	logger.Info("csharp")
	logger.Info("\nRun `gauge --init <template_name>` to create a new Gauge project.")
}
