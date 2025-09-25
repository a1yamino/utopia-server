package node

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"utopia-server/internal/models"
)

type mysqlStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) Store {
	return &mysqlStore{db: db}
}

func (s *mysqlStore) CreateNode(node *models.Node) error {
	gpus, err := json.Marshal(node.Gpus)
	if err != nil {
		return fmt.Errorf("failed to marshal gpus: %w", err)
	}

	query := "INSERT INTO nodes (hostname, status, gpus, control_port, last_seen) VALUES (?, ?, ?, ?, ?)"
	result, err := s.db.Exec(query, node.Hostname, node.Status, gpus, node.ControlPort, node.LastSeen)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	node.ID = id
	return nil
}

func (s *mysqlStore) GetNode(id int64) (*models.Node, error) {
	query := "SELECT id, hostname, status, gpus, control_port, last_seen FROM nodes WHERE id = ?"
	row := s.db.QueryRow(query, id)

	var node models.Node
	var gpus []byte
	err := row.Scan(&node.ID, &node.Hostname, &node.Status, &gpus, &node.ControlPort, &node.LastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found")
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	err = json.Unmarshal(gpus, &node.Gpus)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gpus: %w", err)
	}

	return &node, nil
}

func (s *mysqlStore) ListNodes() ([]*models.Node, error) {
	query := "SELECT id, hostname, status, gpus, control_port, last_seen FROM nodes"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		var gpus []byte
		err := rows.Scan(&node.ID, &node.Hostname, &node.Status, &gpus, &node.ControlPort, &node.LastSeen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		err = json.Unmarshal(gpus, &node.Gpus)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal gpus: %w", err)
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

func (s *mysqlStore) UpdateNode(node *models.Node) error {
	gpus, err := json.Marshal(node.Gpus)
	if err != nil {
		return fmt.Errorf("failed to marshal gpus for update: %w", err)
	}

	query := "UPDATE nodes SET status = ?, control_port = ?, last_seen = ?, gpus = ? WHERE id = ?"
	_, err = s.db.Exec(query, node.Status, node.ControlPort, node.LastSeen, gpus, node.ID)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}
	return nil
}
