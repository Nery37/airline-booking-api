package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"

	"airline-booking/internal/config"
	"airline-booking/internal/models"
)

type Client struct {
	es     *elasticsearch.Client
	logger *zap.Logger
}

type FlightDocument struct {
	ID            int64     `json:"id"`
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	Airline       string    `json:"airline"`
	Aircraft      string    `json:"aircraft"`
	FareClass     string    `json:"fare_class"`
	BasePrice     float64   `json:"base_price"`
}

type HoldDocument struct {
	ID        int64      `json:"id"`
	FlightID  int64      `json:"flight_id"`
	SeatNo    string     `json:"seat_no"`
	HolderID  string     `json:"holder_id"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Status    string     `json:"status"` // "active", "expired", "confirmed"
}

type TicketDocument struct {
	ID          int64     `json:"id"`
	FlightID    int64     `json:"flight_id"`
	SeatNo      string    `json:"seat_no"`
	UserID      string    `json:"user_id"`
	PriceAmount int64     `json:"price_amount"`
	Currency    string    `json:"currency"`
	IssuedAt    time.Time `json:"issued_at"`
	PnrCode     string    `json:"pnr_code"`
	PaymentRef  string    `json:"payment_ref"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"` // "confirmed", "cancelled"
}

type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source FlightDocument `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

const FlightsIndex = "flights"
const HoldsIndex = "holds"
const TicketsIndex = "tickets"

func NewClient(cfg *config.ElasticsearchConfig, logger *zap.Logger) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}

	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	logger.Info("Connected to Elasticsearch successfully")

	return &Client{
		es:     es,
		logger: logger,
	}, nil
}

func (c *Client) CreateIndex(ctx context.Context) error {
	// Create flights index
	flightsMapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"origin": {"type": "keyword"},
				"destination": {"type": "keyword"},
				"departure_time": {"type": "date"},
				"arrival_time": {"type": "date"},
				"airline": {"type": "keyword"},
				"aircraft": {"type": "keyword"},
				"fare_class": {"type": "keyword"},
				"base_price": {"type": "long"}
			}
		}
	}`

	if err := c.createSingleIndex(ctx, FlightsIndex, flightsMapping); err != nil {
		return err
	}

	// Create holds index
	holdsMapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"flight_id": {"type": "long"},
				"seat_no": {"type": "keyword"},
				"holder_id": {"type": "keyword"},
				"expires_at": {"type": "date"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"status": {"type": "keyword"}
			}
		}
	}`

	if err := c.createSingleIndex(ctx, HoldsIndex, holdsMapping); err != nil {
		return err
	}

	// Create tickets index
	ticketsMapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"flight_id": {"type": "long"},
				"seat_no": {"type": "keyword"},
				"user_id": {"type": "keyword"},
				"price_amount": {"type": "long"},
				"currency": {"type": "keyword"},
				"issued_at": {"type": "date"},
				"pnr_code": {"type": "keyword"},
				"payment_ref": {"type": "keyword"},
				"created_at": {"type": "date"},
				"status": {"type": "keyword"}
			}
		}
	}`

	if err := c.createSingleIndex(ctx, TicketsIndex, ticketsMapping); err != nil {
		return err
	}

	return nil
}

func (c *Client) createSingleIndex(ctx context.Context, indexName, mapping string) error {
	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 { // 400 might be "index already exists"
		return fmt.Errorf("failed to create index %s: %s", indexName, res.String())
	}

	c.logger.Info("Elasticsearch index created/verified successfully", zap.String("index", indexName))
	return nil
}

func (c *Client) IndexFlight(ctx context.Context, flight FlightDocument) error {
	body, err := json.Marshal(flight)
	if err != nil {
		return fmt.Errorf("failed to marshal flight: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      FlightsIndex,
		DocumentID: strconv.FormatInt(flight.ID, 10),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to index flight: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index flight: %s", res.String())
	}

	return nil
}

func (c *Client) BulkIndexFlights(ctx context.Context, flights []FlightDocument) error {
	if len(flights) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, flight := range flights {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": FlightsIndex,
				"_id":    strconv.FormatInt(flight.ID, 10),
			},
		}

		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta: %w", err)
		}

		docBytes, err := json.Marshal(flight)
		if err != nil {
			return fmt.Errorf("failed to marshal flight: %w", err)
		}

		buf.Write(metaBytes)
		buf.WriteByte('\n')
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to bulk index flights: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to bulk index flights: %s", res.String())
	}

	c.logger.Info("Bulk indexed flights successfully", zap.Int("count", len(flights)))
	return nil
}

func (c *Client) SearchFlights(ctx context.Context, req models.FlightSearchRequest) (*models.FlightSearchResponse, error) {
	query := c.buildSearchQuery(req)
	
	from := (req.Page - 1) * req.Size
	
	searchBody := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  req.Size,
		"sort": []map[string]interface{}{
			{"departure_time": map[string]interface{}{"order": "asc"}},
		},
	}

	body, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	searchReq := esapi.SearchRequest{
		Index: []string{FlightsIndex},
		Body:  bytes.NewReader(body),
	}

	res, err := searchReq.Do(ctx, c.es)
	if err != nil {
		return nil, fmt.Errorf("failed to search flights: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var searchRes SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchRes); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	flights := make([]models.FlightSearchResult, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		flights[i] = models.FlightSearchResult{
			ID:            hit.Source.ID,
			Origin:        hit.Source.Origin,
			Destination:   hit.Source.Destination,
			DepartureTime: hit.Source.DepartureTime,
			ArrivalTime:   hit.Source.ArrivalTime,
			Airline:       hit.Source.Airline,
			Aircraft:      hit.Source.Aircraft,
			FareClass:     hit.Source.FareClass,
			BasePrice:     hit.Source.BasePrice,
		}
	}

	return &models.FlightSearchResponse{
		Flights: flights,
		Total:   searchRes.Hits.Total.Value,
		Page:    req.Page,
		Size:    req.Size,
	}, nil
}

func (c *Client) buildSearchQuery(req models.FlightSearchRequest) map[string]interface{} {
	must := []map[string]interface{}{
		{"term": map[string]interface{}{"origin": req.Origin}},
		{"term": map[string]interface{}{"destination": req.Destination}},
	}

	// Date range query
	if req.Date != "" {
		startDate := req.Date + "T00:00:00Z"
		endDate := req.Date + "T23:59:59Z"
		
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"departure_time": map[string]interface{}{
					"gte": startDate,
					"lte": endDate,
				},
			},
		})
	}

	// Optional filters
	if req.FareClass != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{"fare_class": req.FareClass},
		})
	}

	if req.Airline != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{"airline": req.Airline},
		})
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must": must,
		},
	}
}

// Hold-related methods
func (c *Client) IndexHold(ctx context.Context, hold HoldDocument) error {
	body, err := json.Marshal(hold)
	if err != nil {
		return fmt.Errorf("failed to marshal hold: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      HoldsIndex,
		DocumentID: strconv.FormatInt(hold.ID, 10),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to index hold: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index hold: %s", res.String())
	}

	c.logger.Info("Hold indexed successfully", zap.Int64("hold_id", hold.ID))
	return nil
}

func (c *Client) UpdateHoldStatus(ctx context.Context, holdID int64, status string) error {
	updateBody := map[string]interface{}{
		"doc": map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	body, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal hold update: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      HoldsIndex,
		DocumentID: strconv.FormatInt(holdID, 10),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to update hold: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to update hold: %s", res.String())
	}

	c.logger.Info("Hold status updated successfully", zap.Int64("hold_id", holdID), zap.String("status", status))
	return nil
}

func (c *Client) DeleteHold(ctx context.Context, holdID int64) error {
	req := esapi.DeleteRequest{
		Index:      HoldsIndex,
		DocumentID: strconv.FormatInt(holdID, 10),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to delete hold: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete hold: %s", res.String())
	}

	c.logger.Info("Hold deleted successfully", zap.Int64("hold_id", holdID))
	return nil
}

// Ticket-related methods
func (c *Client) IndexTicket(ctx context.Context, ticket TicketDocument) error {
	body, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      TicketsIndex,
		DocumentID: strconv.FormatInt(ticket.ID, 10),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to index ticket: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index ticket: %s", res.String())
	}

	c.logger.Info("Ticket indexed successfully", zap.Int64("ticket_id", ticket.ID), zap.String("pnr", ticket.PnrCode))
	return nil
}

func (c *Client) UpdateTicketStatus(ctx context.Context, ticketID int64, status string) error {
	updateBody := map[string]interface{}{
		"doc": map[string]interface{}{
			"status": status,
		},
	}

	body, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket update: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      TicketsIndex,
		DocumentID: strconv.FormatInt(ticketID, 10),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to update ticket: %s", res.String())
	}

	c.logger.Info("Ticket status updated successfully", zap.Int64("ticket_id", ticketID), zap.String("status", status))
	return nil
}

// Utility method to check document count
func (c *Client) GetDocumentCount(ctx context.Context, index string) (int64, error) {
	req := esapi.CountRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("count error: %s", res.String())
	}

	var countRes struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&countRes); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}

	return countRes.Count, nil
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(index string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// CountDocuments returns the document count in an index
func (c *Client) CountDocuments(index string) (int64, error) {
	return c.GetDocumentCount(context.Background(), index)
}

// IndexDocument indexes a document with a given ID
func (c *Client) IndexDocument(index, docID string, doc map[string]interface{}) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

// CreateIndexWithMapping creates an index with a mapping (method for seeder)
func (c *Client) CreateIndexWithMapping(index string, mapping map[string]interface{}) error {
	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	return nil
}
