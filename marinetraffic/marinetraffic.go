package marinetraffic

import (
	"fmt"
	"github.com/erikbryant/web"
)

// ShipsInRegion finds which ships are in a given region
func ShipsInRegion(x, y, zoom string) ([]map[string]string, error) {
	// The area visible from our apartment is
	//   x := "1309"
	//   y := "3165"
	//   zoom := "14"

	url := "https://www.marinetraffic.com/getData/get_data_json_4/z:" + zoom + "/X:" + x + "/Y:" + y + "/station:0"
	headers := map[string]string{
		"user-agent":       "ship-ahoy",
		"x-requested-with": "XMLHttpRequest",
		"vessel-image":     "001609ab6d06a620f459d4a1fd65f1315f11",
	}

	region, err := web.RequestJSON(url, headers)
	if err != nil {
		return nil, err
	}

	if web.ToInt(region["type"]) != 1 {
		return nil, fmt.Errorf("Unexpected response type %v", region)
	}

	data := region["data"].(map[string]interface{})
	rows := data["rows"].([]interface{})
	areaShips := web.ToInt(data["areaShips"])
	if len(rows) != areaShips {
		return nil, fmt.Errorf("Number of ships returned did not match expected %d %v", areaShips, region)
	}

	// Convert from nested interfaces to defined types.
	var rowMap []map[string]string
	for _, row := range rows {
		newRow := make(map[string]string)
		for key, value := range row.(map[string]interface{}) {
			newRow[key] = web.ToString(value)
		}
		rowMap = append(rowMap, newRow)
	}

	return rowMap, nil
}
