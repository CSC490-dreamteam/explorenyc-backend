package solver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// RouteSolver sends a SolverInput to a routing engine and gets back a SolverOutput.
type RouteSolver interface {
	Solve(input SolverInput) (SolverOutput, error)
}

// UnprocessableInputError indicates the solver was given constraints it
// could not satisfy (HTTP 422), as opposed to a transport/server failure.
type UnprocessableInputError struct {
	Body []byte
}

func (e *UnprocessableInputError) Error() string {
	return fmt.Sprintf("solver rejected input as unprocessable: %s", string(e.Body))
}

// PythonSolver calls the python route-generation microservice.
type PythonSolver struct {
	URL    string
	Client *http.Client
}

func NewPythonSolver(url string) PythonSolver {
	return PythonSolver{
		URL:    url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p PythonSolver) Solve(input SolverInput) (SolverOutput, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return SolverOutput{}, fmt.Errorf("failed to marshal solver input: %w", err)
	}

	req, err := http.NewRequest("POST", p.URL, bytes.NewBuffer(payload))
	if err != nil {
		return SolverOutput{}, fmt.Errorf("failed to create solver request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return SolverOutput{}, fmt.Errorf("failed to reach solver service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		return SolverOutput{}, &UnprocessableInputError{Body: body}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return SolverOutput{}, fmt.Errorf("solver service returned %d: %s", resp.StatusCode, string(body))
	}

	var output SolverOutput
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return SolverOutput{}, fmt.Errorf("failed to decode solver response: %w", err)
	}

	return output, nil
}
