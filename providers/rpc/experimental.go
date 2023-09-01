package rpc

import (
	"github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
)

// embedPayload will embed the RequestPayload into the given JSON object at the dot path notation location ("object.data").
func (p *RequestPayload) embedPayload(rawJSON []byte, dotPath string) ([]byte, error) {
	if rawJSON != nil {
		jdata2, err := yaml.YAMLToJSON(rawJSON)
		if err != nil {
			return nil, err
		}
		g, err := gabs.ParseJSON(jdata2)
		if err != nil {
			return nil, err
		}
		if _, err := g.SetP(p, dotPath); err != nil {
			return nil, err
		}

		return g.Bytes(), nil
	}

	return rawJSON, nil
}
