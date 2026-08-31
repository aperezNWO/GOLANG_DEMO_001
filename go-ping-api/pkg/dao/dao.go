package dao

import (
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

// With this line:
const connString = "server=webapiangulardemo.mssql.somee.com;port=1433;user id=aperezNWO_SQLLogin_1;password=aperezNWO_SQLLogin_1;database=webapiangulardemo;encrypt=false;TrustServerCertificate=true"

// AccessLog mirrors AccessLog entity struct
type AccessLog struct {
	IDColumn   int64   `json:"id_Column"`
	PageName   *string `json:"pageName"`
	AccessDate *string `json:"accessDate"`
	IPValue    *string `json:"ipValue"`
}

// PersonaTable mirrors PersonaTable entity struct
type PersonaTable struct {
	IDColumn       int64   `json:"id_Column"`
	Ciudad         *string `json:"ciudad"`
	NombreCompleto *string `json:"nombreCompleto"`
}

type DAOManager struct{}

func NewDAOManager() *DAOManager {
	return &DAOManager{}
}

func (d *DAOManager) getConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}
	return db, nil
}

func (d *DAOManager) GetAllLogs() ([]AccessLog, error) {
	db, err := d.getConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
        SELECT TOP 100
               AL.[ID_column]     AS id_column
             , AL.[PageName]      AS pageName
             , AL.[AccessDate]    AS accessDate
             , AL.[IpValue]       AS ipValue
        FROM
            dbo.accessLogs AL
        WHERE
            AL.[LogType] = 1
        AND
            (AL.PAGENAME LIKE '%DEMO%'
        AND
            AL.PAGENAME LIKE '%PAGE%')
        AND
            AL.PAGENAME NOT LIKE '%ERROR%'
        AND
            AL.PAGENAME NOT LIKE '%PAGE_DEMO_INDEX%'
        AND
            UPPER(AL.PAGENAME) NOT LIKE '%CACHE%'
        AND
            AL.IPVALUE <> '::1'
        ORDER BY
            AL.[ID_column] DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AccessLog
	for rows.Next() {
		var logEntry AccessLog
		err := rows.Scan(&logEntry.IDColumn, &logEntry.PageName, &logEntry.AccessDate, &logEntry.IPValue)
		if err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}

func (d *DAOManager) GetAllPersons() ([]PersonaTable, error) {
	db, err := d.getConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
        SELECT
            [Id_Column]            AS id_column
            ,[Ciudad]              AS ciudad
            ,[NombreCompleto]      AS nombreCompleto
        FROM
            [dbo].[Persona]
        ORDER BY
            Id_Column ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var personas []PersonaTable
	for rows.Next() {
		var person PersonaTable
		err := rows.Scan(&person.IDColumn, &person.Ciudad, &person.NombreCompleto)
		if err != nil {
			return nil, err
		}
		personas = append(personas, person)
	}

	return personas, nil
}	