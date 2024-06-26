package busservice

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sarcb/catalyst-sp24/bus"
	"github.com/sarcb/catalyst-sp24/generated/model"
)

func (h *busService) handleJob(automationMsg *bus.JobMsg) {
	ctx := busContext()

	job, err := h.db.JobCreate(ctx, automationMsg.ID, &model.JobForm{
		Automation: automationMsg.Automation,
		Payload:    automationMsg.Message.Payload,
		Origin:     automationMsg.Origin,
	})
	if err != nil {
		log.Println(err)

		return
	}

	automation, err := h.db.AutomationGet(ctx, automationMsg.Automation)
	if err != nil {
		log.Println(err)

		return
	}

	if automation.Script == "" {
		log.Println("automation is empty")

		return
	}

	if automationMsg.Message.Secrets == nil {
		automationMsg.Message.Secrets = map[string]string{}
	}
	automationMsg.Message.Secrets["catalyst_apikey"] = h.apiKey
	automationMsg.Message.Secrets["catalyst_apiurl"] = h.apiURL

	scriptMessage, err := json.Marshal(automationMsg.Message)
	if err != nil {
		log.Println(err)

		return
	}

	containerID, logs, err := createContainer(ctx, automation.Image, automation.Script, string(scriptMessage), h.network)
	if err != nil {
		log.Println(err)

		return
	}

	if _, err := h.db.JobUpdate(ctx, automationMsg.ID, &model.JobUpdate{
		Container: &containerID,
		Running:   true,
		Output:    job.Output,
		Log:       &logs,
		Status:    job.Status,
	}); err != nil {
		log.Println(err)

		return
	}

	var result map[string]any

	stdout, _, err := runDocker(ctx, automationMsg.ID, containerID, h.db)
	if err != nil {
		result = map[string]any{"error": fmt.Sprintf("error running script %s %s", err, string(stdout))}
	} else {
		var data map[string]any
		if err := json.Unmarshal(stdout, &data); err != nil {
			result = map[string]any{"error": string(stdout)}
		} else {
			result = data
		}
	}

	h.catalystBus.ResultChannel.Publish(&bus.ResultMsg{Automation: automationMsg.Automation, Data: result, Target: automationMsg.Origin})

	if err := h.db.JobComplete(ctx, automationMsg.ID, result); err != nil {
		log.Println(err)

		return
	}
}

/*
func getAutomation(automationID string, config *Config) (*model.AutomationResponse, error) {
	req, err := http.NewRequest(http.MethodGet, config.CatalystAPIUrl+"/automations/"+automationID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", config.CatalystAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var automation model.AutomationResponse
	if err := json.Unmarshal(b, &automation); err != nil {
		return nil, err
	}
	return &automation, nil
}
*/
