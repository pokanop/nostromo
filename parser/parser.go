package parser

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/pokanop/nostromo/model"
)

type Parser struct{}

func (p *Parser) Parse(data io.Reader) (*model.Manifest, error) {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}

	var m *model.Manifest
	err = json.Unmarshal(b, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
