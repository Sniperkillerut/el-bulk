package main

import (
	"encoding/json"
	"fmt"
	"github.com/el-bulk/backend/models"
)

func main() {
	p := models.Product{
		ID:           "test-1",
		Name:         "Black Lotus",
		CostBasisCOP: 1000000.0,
	}

	b, _ := json.Marshal(p)
	fmt.Println(string(b))
}
