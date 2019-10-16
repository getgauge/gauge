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
		logger.Fatalf(true, "Gauge templates URL %s is not reachable: %s", templatesURL, err.Error())
	}
	res, err := http.Get(templatesURL)
	if err != nil {
		logger.Fatalf(true, "Error occurred while downloading templates list from %s: %s", templatesURL, err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		logger.Errorf(true, fmt.Sprintf("Error downloading templates info: %s.\n%s", templatesURL, res.Status))
	}
	templates := []template{}
	err = json.NewDecoder(res.Body).Decode(&templates)
	if err != nil {
		logger.Errorf(true, fmt.Sprintf("Error decoding templates info: %s : %s", templatesURL, err.Error()))
	}
	for _, t := range templates {
		logger.Infof(true, t.GetName())
	}
	logger.Infof(true, "csharp")
	logger.Infof(true, "\nRun `gauge init <template_name>` to create a new Gauge project.")
}
