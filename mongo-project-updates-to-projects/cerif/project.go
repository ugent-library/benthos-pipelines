package cerif

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
)

var ErrNonCompliantXml = errors.New("non compliant cerif xml")

type Project struct {
	ID             string                      `json:"id,omitempty"`
	StartDate      string                      `json:"start_date,omitempty"`
	EndDate        string                      `json:"end_date,omitempty"`
	Acronym        string                      `json:"acronym,omitempty"`
	Title          []TranslatedString          `json:"title,omitempty"`
	Abstract       []TranslatedString          `json:"abstract,omitempty"`
	Keyword        []TranslatedString          `json:"keyword,omitempty"`
	Classification []Classification            `json:"classfication,omitempty"`
	FederatedIDS   []FederatedIDClassification `json:"federated_ids,omitempty"`
	SentDateTime   time.Time                   `json:"sent_date_time,omitempty"`
	Action         string                      `json:"action,omitempty"`
}

type TranslatedString struct {
	Translation string `json:"translation,omitempty"`
	Lang        string `json:"lang,omitempty"`
	Value       string `json:"value,omitempty"`
}

type Classification struct {
	URI         string               `json:"uri,omitempty"`
	Name        []TranslatedString   `json:"name,omitempty"`
	Description []TranslatedString   `json:"description,omitempty"`
	Terms       []ClassificationTerm `json:"terms,omitempty"`
}

type ClassificationTerm struct {
	URI       string             `json:"uri,omitempty"`
	Term      []TranslatedString `json:"term,omitempty"`
	StartDate time.Time          `json:"start_date,omitempty"`
	EndDate   time.Time          `json:"end_date,omitempty"`
}

type FederatedIDClassification struct {
	URI         string             `json:"uri,omitempty"`
	Name        []TranslatedString `json:"name,omitempty"`
	Description []TranslatedString `json:"description,omitempty"`
	IDS         []FederatedID      `json:"ids,omitempty"`
}

type FederatedID struct {
	ID        string             `json:"id,omitempty"`
	URI       string             `json:"uri,omitempty"`
	Term      []TranslatedString `json:"term,omitempty"`
	StartDate time.Time          `json:"start_date,omitempty"`
	EndDate   time.Time          `json:"end_date,omitempty"`
}

func ParseProject(buf []byte) (*Project, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	node := xmlquery.FindOne(doc, "//ns0:cfProj")
	if node == nil {
		return nil, fmt.Errorf("not a valid cerif document")
	}

	projects := xmlquery.FindOne(doc, "//ns0:projects")
	if projects == nil {
		return nil, fmt.Errorf("not a valid cerif document")
	}

	p := &Project{}

	for _, n := range node.SelectElements("*") {
		// log.Println(n.Data)
		switch n.Data {
		case "cfProjId":
			p.ID = n.InnerText()
		case "cfStartDate":
			p.StartDate = n.InnerText()
		case "cfEndDate":
			p.EndDate = n.InnerText()
		case "cfAcro":
			p.Acronym = n.InnerText()
		case "cfTitle":
			p.Title = parseTranslatedField(p.Title, n)
		case "cfAbstr":
			p.Abstract = parseTranslatedField(p.Abstract, n)
		case "cfKeyw":
			p.Keyword = parseTranslatedField(p.Keyword, n)
		case "cfProj_Class":
			p.Classification = parseClassification(p.Classification, n, doc)
		case "cfFedId":
			p.FederatedIDS = parseFederatedIDClassification(p.FederatedIDS, n, doc)
		}
	}

	if val, err := parseDateTimeField(projects.SelectAttr("sentDateTime")); err == nil {
		p.SentDateTime = val
	}

	if action := node.SelectAttr("action"); action != "" {
		p.Action = action
	} else {
		p.Action = "UPDATE"
	}

	return p, nil
}

func parseDateTimeField(dt string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(dt))
}

func parseTranslatedField(f []TranslatedString, n *xmlquery.Node) []TranslatedString {
	s := TranslatedString{
		Translation: n.SelectAttr("cfTrans"),
		Lang:        n.SelectAttr("cfLangCode"),
		Value:       n.InnerText(),
	}

	if f == nil {
		f = []TranslatedString{s}
	} else {
		f = append(f, s)
	}

	return f
}

func parseClassification(field []Classification, term *xmlquery.Node, doc *xmlquery.Node) []Classification {
	// Resolve the classification for the term
	classSchemeId := xmlquery.FindOne(term, "/ns1:cfClassSchemeId").InnerText()
	path := fmt.Sprintf("//ns0:cfClassScheme/ns1:cfClassSchemeId[.='%s']/..", classSchemeId)
	classScheme := xmlquery.FindOne(doc, path)

	tmp := Classification{}
	for _, v := range classScheme.SelectElements("*") {
		switch v.Data {
		case "cfURI":
			tmp.URI = v.InnerText()
		case "cfDescr":
			tmp.Description = parseTranslatedField(tmp.Description, v)
		case "cfName":
			tmp.Name = parseTranslatedField(tmp.Name, v)
		}
	}

	// Fetch if classification already was added to project
	var classification *Classification
	for k, v := range field {
		if v.URI == tmp.URI {
			classification = &field[k]
			break
		}
	}

	if classification == nil {
		field = append(field, tmp)
		classification = &tmp
	}

	// Transform term to a ClassificationTerm
	t := ClassificationTerm{}

	// Resolve the term for the term id
	classId := xmlquery.FindOne(term, "/ns1:cfClassId").InnerText()
	path = fmt.Sprintf("//ns0:cfClass/ns1:cfClassId[.='%s']/..", classId)
	class := xmlquery.FindOne(doc, path)

	for _, v := range term.SelectElements("*") {
		switch v.Data {
		case "cfStartDate":
			if val, err := parseDateTimeField(v.InnerText()); err == nil {
				t.StartDate = val
			}
		case "cfEndDate":
			if val, err := parseDateTimeField(v.InnerText()); err == nil {
				t.EndDate = val
			}
		}
	}

	for _, v := range class.SelectElements("*") {
		switch v.Data {
		case "cfURI":
			t.URI = v.InnerText()
		case "cfTerm":
			t.Term = parseTranslatedField(t.Term, v)
		}
	}

	// Add or replace ClassificationTerm to / in Classification
	found := false
	for k, v := range classification.Terms {
		if v.URI == t.URI {
			found = true
			classification.Terms[k] = t
			break
		}
	}

	if !found {
		classification.Terms = append(classification.Terms, t)
	}

	// Update the field & return the entire value
	for k, cc := range field {
		if cc.URI == classification.URI {
			field[k] = *classification
		}
	}

	return field
}

func parseFederatedIDClassification(field []FederatedIDClassification, term *xmlquery.Node, doc *xmlquery.Node) []FederatedIDClassification {
	// Resolve the classification for the term
	classSchemeId := xmlquery.FindOne(term, "/ns1:cfClassSchemeId").InnerText()
	path := fmt.Sprintf("//ns0:cfClassScheme/ns1:cfClassSchemeId[.='%s']/..", classSchemeId)
	classScheme := xmlquery.FindOne(doc, path)

	tmp := FederatedIDClassification{}
	for _, v := range classScheme.SelectElements("*") {
		//log.Println(v.Data)
		switch v.Data {
		case "cfURI":
			tmp.URI = v.InnerText()
		case "cfDescr":
			tmp.Description = parseTranslatedField(tmp.Description, v)
		case "cfName":
			tmp.Name = parseTranslatedField(tmp.Name, v)
		}
	}

	// Fetch if classification already was added to project
	var fedIdclassification *FederatedIDClassification
	for k, v := range field {
		if v.URI == tmp.URI {
			fedIdclassification = &field[k]
			break
		}
	}

	if fedIdclassification == nil {
		field = append(field, tmp)
		fedIdclassification = &tmp
	}

	// Transform term to a FederatedID
	c := FederatedID{}

	// Resolve the term for the term id
	classId := xmlquery.FindOne(term, "/ns1:cfClassId").InnerText()
	path = fmt.Sprintf("//ns0:cfClass/ns1:cfClassId[.='%s']/..", classId)
	class := xmlquery.FindOne(doc, path)

	for _, v := range term.SelectElements("*") {
		// log.Printf("** %s", v.Data)
		switch v.Data {
		case "cfFedId":
			c.ID = v.InnerText()
		case "cfStartDate":
			if val, err := parseDateTimeField(v.InnerText()); err == nil {
				c.StartDate = val
			}
		case "cfEndDate":
			if val, err := parseDateTimeField(v.InnerText()); err == nil {
				c.EndDate = val
			}
		}
	}

	for _, v := range class.SelectElements("*") {
		// log.Printf("-- %s", v.Data)
		switch v.Data {
		case "cfURI":
			c.URI = v.InnerText()
		case "cfTerm":
			c.Term = parseTranslatedField(c.Term, v)
		}
	}

	// Add or replace FederatedID to / in FederatedIDClassification
	found := false
	for k, v := range fedIdclassification.IDS {
		if v.ID == c.ID && v.URI == c.URI {
			found = true
			fedIdclassification.IDS[k] = c
			break
		}
	}

	if !found {
		fedIdclassification.IDS = append(fedIdclassification.IDS, c)
	}

	// Update the field & return the entire value
	for k, cc := range field {
		if cc.URI == fedIdclassification.URI {
			field[k] = *fedIdclassification
		}
	}

	return field
}
