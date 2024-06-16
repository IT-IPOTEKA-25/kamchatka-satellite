package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"math"
	"strings"
	"time"
	"unicode"
)

type Server struct {
	conn *pgx.Conn
}

func NewServer(conn *pgx.Conn) *Server {
	return &Server{
		conn: conn,
	}
}

func parseDMSString(dmsString string) (int, int, int, error) {
	dmsString = strings.TrimLeftFunc(dmsString, func(r rune) bool {
		return unicode.IsLetter(r)
	})
	var degrees, minutes, seconds int
	_, err := fmt.Sscanf(dmsString, "%d°%d'%d\"", &degrees, &minutes, &seconds)
	if err != nil {
		return 0, 0, 0, err
	}
	return degrees, minutes, seconds, nil
}

func convertDMSToDD(data string) float32 {
	degrees, minutes, seconds, _ := parseDMSString(data)
	decimalDegrees := float64(degrees) + float64(minutes)/60.0 + float64(seconds)/3600.0
	return float32(math.Round(decimalDegrees*10000) / 10000) // округление до 4 знаков после запятой
}

type CoordinatesDB struct {
	Name string
	Dot  []string
}

type Coordinates struct {
	Name string
	Dot  []float32
}

func (s *Server) GetRouteCoordinates(ctx context.Context) ([]*Coordinates, error) {
	var dbResult string
	var coordinatesDB [][]CoordinatesDB
	err := s.conn.QueryRow(ctx, "select coordinates from kamchatka.security_territories_coordinates").Scan(&dbResult)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(dbResult), &coordinatesDB)
	if err != nil {
		return nil, err
	}
	var coordinates []*Coordinates
	for id := range coordinatesDB {
		for _, coordinate := range coordinatesDB[id] {
			coordinates = append(coordinates, &Coordinates{
				Name: coordinate.Name,
				Dot: []float32{
					convertDMSToDD(coordinate.Dot[0]),
					convertDMSToDD(coordinate.Dot[1]),
				},
			})
		}
	}

	return coordinates, nil
}

func (s *Server) AddSatelliteAlert(ctx context.Context, image string, category string, x1 string, y1 string, x2 string, y2 string) error {
	_, sqlErr := s.conn.Exec(ctx, "insert into kamchatka.satellite_alerts(image, category, time, coordinates, handled) VALUES ($1, $2, $3, $4, $5)", image, category, time.Now(), fmt.Sprintf("[\"%s\", \"%s\", \"%s\", \"%s\"]", x1, y1, x2, y2), false)
	if sqlErr != nil {
		return sqlErr
	}
	return nil
}
