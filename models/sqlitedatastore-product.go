package models

import (
	"database/sql"
	"reflect"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // register sqlite3 driver
	log "github.com/sirupsen/logrus"
	"github.com/tbellembois/gochimitheque/constants"
	"github.com/tbellembois/gochimitheque/helpers"
)

// IsProductBookmark returns true if there is a bookmark for the product pr for the person pe
func (db *SQLiteDataStore) IsProductBookmark(pr Product, pe Person) (bool, error) {
	var (
		sqlr string
		err  error
		i    int
	)
	sqlr = `SELECT count(*) FROM bookmark WHERE person = ? AND product = ?`
	if err = db.Get(&i, sqlr, pe.PersonID, pr.ProductID); err != nil {
		return false, err
	}
	return i != 0, err
}

// CreateProductBookmark bookmarks the product pr for the person pe
func (db *SQLiteDataStore) CreateProductBookmark(pr Product, pe Person) error {
	var (
		sqlr string
		err  error
	)
	sqlr = `INSERT into bookmark(person, product) VALUES (? , ?)`
	if _, err = db.Exec(sqlr, pe.PersonID, pr.ProductID); err != nil {
		return err
	}
	return nil
}

// DeleteProductBookmark remove the bookmark for the product pr and the person pe
func (db *SQLiteDataStore) DeleteProductBookmark(pr Product, pe Person) error {
	var (
		sqlr string
		err  error
	)
	sqlr = `DELETE from bookmark WHERE person = ? AND product = ?`
	if _, err = db.Exec(sqlr, pe.PersonID, pr.ProductID); err != nil {
		return err
	}
	return nil
}

// GetProductsCasNumbers return the cas numbers matching the search criteria
func (db *SQLiteDataStore) GetProductsCasNumbers(p helpers.Dbselectparam) ([]CasNumber, int, error) {
	var (
		casnumbers                         []CasNumber
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT casnumber.casnumber_id)")
	presreq.WriteString(" SELECT casnumber_id, casnumber_label")

	comreq.WriteString(" FROM casnumber")
	comreq.WriteString(" WHERE casnumber_label LIKE :search")
	postsreq.WriteString(" ORDER BY casnumber_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&casnumbers, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	// setting the C attribute for formula matching exactly the search
	s := p.GetSearch()
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimSuffix(s, "%")
	var casn CasNumber

	r := db.QueryRowx(`SELECT casnumber_id, casnumber_label FROM casnumber WHERE casnumber_label == ?`, s)
	if err = r.StructScan(&casn); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	} else {
		for i, c := range casnumbers {
			if c.CasNumberID == casn.CasNumberID {
				casnumbers[i].C = 1
			}
		}
	}

	log.WithFields(log.Fields{"casnumbers": casnumbers}).Debug("GetProductsCasNumbers")
	return casnumbers, count, nil
}

// GetProductsCeNumbers return the cas numbers matching the search criteria
func (db *SQLiteDataStore) GetProductsCeNumbers(p helpers.Dbselectparam) ([]CeNumber, int, error) {
	var (
		cenumbers                          []CeNumber
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT cenumber.cenumber_id)")
	presreq.WriteString(" SELECT cenumber_id, cenumber_label")

	comreq.WriteString(" FROM cenumber")
	comreq.WriteString(" WHERE cenumber_label LIKE :search")
	postsreq.WriteString(" ORDER BY cenumber_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&cenumbers, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	// setting the C attribute for formula matching exactly the search
	s := p.GetSearch()
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimSuffix(s, "%")
	var cen CeNumber

	r := db.QueryRowx(`SELECT cenumber_id, cenumber_label FROM cenumber WHERE cenumber_label == ?`, s)
	if err = r.StructScan(&cen); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	} else {
		for i, c := range cenumbers {
			if c.CeNumberID == cen.CeNumberID {
				cenumbers[i].C = 1
			}
		}
	}

	log.WithFields(log.Fields{"cenumbers": cenumbers}).Debug("GetProductsCeNumbers")
	return cenumbers, count, nil
}

// GetProductsEmpiricalFormulas return the empirical formulas matching the search criteria
func (db *SQLiteDataStore) GetProductsEmpiricalFormulas(p helpers.Dbselectparam) ([]EmpiricalFormula, int, error) {
	var (
		eformulas                          []EmpiricalFormula
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT empiricalformula.empiricalformula_id)")
	presreq.WriteString(" SELECT empiricalformula_id, empiricalformula_label")

	comreq.WriteString(" FROM empiricalformula")
	comreq.WriteString(" WHERE empiricalformula_label LIKE :search")
	postsreq.WriteString(" ORDER BY empiricalformula_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&eformulas, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	// setting the C attribute for formula matching exactly the search
	s := p.GetSearch()
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimSuffix(s, "%")
	var ef EmpiricalFormula

	r := db.QueryRowx(`SELECT empiricalformula_id, empiricalformula_label FROM empiricalformula WHERE empiricalformula_label == ?`, s)
	if err = r.StructScan(&ef); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	} else {
		for i, e := range eformulas {
			if e.EmpiricalFormulaID == ef.EmpiricalFormulaID {
				eformulas[i].C = 1
			}
		}
	}

	log.WithFields(log.Fields{"eformulas": eformulas}).Debug("GetProductsEmpiricalFormulas")
	return eformulas, count, nil
}

// GetProductsClassOfCompounds return the classe of compounds matching the search criteria
func (db *SQLiteDataStore) GetProductsClassOfCompounds(p helpers.Dbselectparam) ([]ClassOfCompound, int, error) {
	var (
		classofcompounds                   []ClassOfCompound
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT classofcompound.classofcompound_id)")
	presreq.WriteString(" SELECT classofcompound_id, classofcompound_label")

	comreq.WriteString(" FROM classofcompound")
	comreq.WriteString(" WHERE classofcompound_label LIKE :search")
	postsreq.WriteString(" ORDER BY classofcompound_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&classofcompounds, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	// setting the C attribute for formula matching exactly the search
	s := p.GetSearch()
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimSuffix(s, "%")
	var coc ClassOfCompound

	r := db.QueryRowx(`SELECT classofcompound_id, classofcompound_label FROM classofcompound WHERE classofcompound_label == ?`, s)
	if err = r.StructScan(&coc); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	} else {
		for i, e := range classofcompounds {
			if e.ClassOfCompoundID == coc.ClassOfCompoundID {
				classofcompounds[i].C = 1
			}
		}
	}

	log.WithFields(log.Fields{"classofcompounds": classofcompounds}).Debug("GetProductsClassOfCompounds")
	return classofcompounds, count, nil
}

// GetProductsNames return the names matching the search criteria
func (db *SQLiteDataStore) GetProductsNames(p helpers.Dbselectparam) ([]Name, int, error) {
	var (
		names                              []Name
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT name.name_id)")
	presreq.WriteString(" SELECT name_id, name_label")

	comreq.WriteString(" FROM name")
	comreq.WriteString(" WHERE name_label LIKE :search")
	postsreq.WriteString(" ORDER BY name_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&names, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	// setting the C attribute for formula matching exactly the search
	s := p.GetSearch()
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimSuffix(s, "%")
	var name Name

	r := db.QueryRowx(`SELECT name_id, name_label FROM name WHERE name_label == ?`, s)
	if err = r.StructScan(&name); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	} else {
		for i, n := range names {
			if n.NameID == name.NameID {
				names[i].C = 1
			}
		}
	}

	log.WithFields(log.Fields{"names": names}).Debug("GetProductsNames")
	return names, count, nil
}

// GetProductsSymbols return the symbols matching the search criteria
func (db *SQLiteDataStore) GetProductsSymbols(p helpers.Dbselectparam) ([]Symbol, int, error) {
	var (
		symbols                            []Symbol
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT symbol.symbol_id)")
	presreq.WriteString(" SELECT symbol_id, symbol_label, symbol_image")

	comreq.WriteString(" FROM symbol")
	comreq.WriteString(" WHERE symbol_label LIKE :search")
	postsreq.WriteString(" ORDER BY symbol_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&symbols, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	log.WithFields(log.Fields{"symbols": symbols}).Debug("GetProductsSymbols")
	return symbols, count, nil
}

// GetProductsHazardStatements return the hazard statements matching the search criteria
func (db *SQLiteDataStore) GetProductsHazardStatements(p helpers.Dbselectparam) ([]HazardStatement, int, error) {
	var (
		hazardstatements                   []HazardStatement
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT hazardstatement.hazardstatement_id)")
	presreq.WriteString(" SELECT hazardstatement_id, hazardstatement_label, hazardstatement_reference")

	comreq.WriteString(" FROM hazardstatement")
	comreq.WriteString(" WHERE hazardstatement_reference LIKE :search")
	postsreq.WriteString(" ORDER BY hazardstatement_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&hazardstatements, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	log.WithFields(log.Fields{"hazardstatements": hazardstatements}).Debug("GetProductsHazardStatements")
	return hazardstatements, count, nil
}

// GetProductsPrecautionaryStatements return the hazard statements matching the search criteria
func (db *SQLiteDataStore) GetProductsPrecautionaryStatements(p helpers.Dbselectparam) ([]PrecautionaryStatement, int, error) {
	var (
		precautionarystatements            []PrecautionaryStatement
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT precautionarystatement.precautionarystatement_id)")
	presreq.WriteString(" SELECT precautionarystatement_id, precautionarystatement_label, precautionarystatement_reference")

	comreq.WriteString(" FROM precautionarystatement")
	comreq.WriteString(" WHERE precautionarystatement_reference LIKE :search")
	postsreq.WriteString(" ORDER BY precautionarystatement_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&precautionarystatements, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	log.WithFields(log.Fields{"precautionarystatements": precautionarystatements}).Debug("GetProductsPrecautionaryStatements")
	return precautionarystatements, count, nil
}

// GetProductsPhysicalStates return the physical states matching the search criteria
func (db *SQLiteDataStore) GetProductsPhysicalStates(p helpers.Dbselectparam) ([]PhysicalState, int, error) {
	var (
		physicalstates                     []PhysicalState
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT physicalstate.physicalstate_id)")
	presreq.WriteString(" SELECT physicalstate_id, physicalstate_label")

	comreq.WriteString(" FROM physicalstate")
	comreq.WriteString(" WHERE physicalstate_label LIKE :search")
	postsreq.WriteString(" ORDER BY physicalstate_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&physicalstates, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	log.WithFields(log.Fields{"physicalstates": physicalstates}).Debug("GetProductsPhysicalStates")
	return physicalstates, count, nil
}

// GetProductsSignalWords return the signal words matching the search criteria
func (db *SQLiteDataStore) GetProductsSignalWords(p helpers.Dbselectparam) ([]SignalWord, int, error) {
	var (
		signalwords                        []SignalWord
		count                              int
		precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                             *sqlx.NamedStmt
		snstmt                             *sqlx.NamedStmt
		err                                error
	)

	precreq.WriteString(" SELECT count(DISTINCT signalword.signalword_id)")
	presreq.WriteString(" SELECT signalword_id, signalword_label")

	comreq.WriteString(" FROM signalword")
	comreq.WriteString(" WHERE signalword_label LIKE :search")
	postsreq.WriteString(" ORDER BY signalword_label  " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search": p.GetSearch(),
		"order":  p.GetOrder(),
		"limit":  p.GetLimit(),
		"offset": p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&signalwords, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	log.WithFields(log.Fields{"signalwords": signalwords}).Debug("GetProductsSignalWords")
	return signalwords, count, nil
}

// GetProducts return the products matching the search criteria
func (db *SQLiteDataStore) GetProducts(p helpers.DbselectparamProduct) ([]Product, int, error) {
	var (
		products                                []Product
		count                                   int
		req, precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                                  *sqlx.NamedStmt
		snstmt                                  *sqlx.NamedStmt
		err                                     error
	)
	log.WithFields(log.Fields{"p": p}).Debug("GetProducts")

	// pre request: select or count
	precreq.WriteString(" SELECT count(DISTINCT p.product_id)")
	presreq.WriteString(` SELECT p.product_id, 
	p.product_specificity, 
	p.product_msds,
	p.product_restricted,
	p.product_radioactive,
	p.product_linearformula,
	p.product_threedformula,
	p.product_disposalcomment,
	p.product_remark,
	empiricalformula.empiricalformula_id AS "empiricalformula.empiricalformula_id",
	empiricalformula.empiricalformula_label AS "empiricalformula.empiricalformula_label",
	physicalstate.physicalstate_id AS "physicalstate.physicalstate_id",
	physicalstate.physicalstate_label AS "physicalstate.physicalstate_label",
	signalword.signalword_id AS "signalword.signalword_id",
	signalword.signalword_label AS "signalword.signalword_label",
	classofcompound.classofcompound_id AS "classofcompound.classofcompound_id",
	classofcompound.classofcompound_label AS "classofcompound.classofcompound_label",
	person.person_id AS "person.person_id",
	person.person_email AS "person.person_email",
	name.name_id AS "name.name_id",
	name.name_label AS "name.name_label",
	bookmark.bookmark_id AS "bookmark.bookmark_id",
	cenumber.cenumber_id AS "cenumber.cenumber_id",
	cenumber.cenumber_label AS "cenumber.cenumber_label",
	casnumber.casnumber_id AS "casnumber.casnumber_id",
	casnumber.casnumber_label AS "casnumber.casnumber_label"`)

	// common parts
	comreq.WriteString(" FROM product as p")
	// get name
	comreq.WriteString(" JOIN name ON p.name = name.name_id")
	// get casnumber
	comreq.WriteString(" JOIN casnumber ON p.casnumber = casnumber.casnumber_id")
	// get cenumber
	comreq.WriteString(" LEFT JOIN cenumber ON p.cenumber = cenumber.cenumber_id")
	// get person
	comreq.WriteString(" JOIN person ON p.person = person.person_id")
	// get physical state
	comreq.WriteString(" LEFT JOIN physicalstate ON p.physicalstate = physicalstate.physicalstate_id")
	// get signal word
	comreq.WriteString(" LEFT JOIN signalword ON p.signalword = signalword.signalword_id")
	// get class of compound
	comreq.WriteString(" LEFT JOIN classofcompound ON p.classofcompound = classofcompound.classofcompound_id")
	// get empirical formula
	comreq.WriteString(" JOIN empiricalformula ON p.empiricalformula = empiricalformula.empiricalformula_id")
	// get bookmark
	comreq.WriteString(" LEFT JOIN bookmark ON (bookmark.product = p.product_id AND bookmark.person = :personid)")
	// get storages, store locations and entities
	if p.GetEntity() != -1 || p.GetStorelocation() != -1 {
		comreq.WriteString(" JOIN storage ON storage.product = p.product_id")
		comreq.WriteString(" JOIN storelocation ON storage.storelocation = storelocation.storelocation_id")
		comreq.WriteString(" JOIN entity ON storelocation.entity = entity.entity_id")
	}
	// get bookmarks
	if p.GetBookmark() {
		comreq.WriteString(" JOIN bookmark AS b ON b.product = p.product_id AND b.person = :personid")
	}
	// filter by permissions
	comreq.WriteString(` JOIN permission AS perm, entity as e ON
	(perm.person = :personid and perm.permission_item_name = "all" and perm.permission_perm_name = "all" and perm.permission_entity_id = e.entity_id) OR
	(perm.person = :personid and perm.permission_item_name = "all" and perm.permission_perm_name = "all" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "all" and perm.permission_perm_name = "r" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "products" and perm.permission_perm_name = "all" and perm.permission_entity_id = e.entity_id) OR
	(perm.person = :personid and perm.permission_item_name = "products" and perm.permission_perm_name = "all" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "products" and perm.permission_perm_name = "r" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "products" and perm.permission_perm_name = "r" and perm.permission_entity_id = e.entity_id)
	`)
	comreq.WriteString(" WHERE name.name_label LIKE :search")
	if p.GetProduct() != -1 {
		comreq.WriteString(" AND p.product_id = :product")
	}
	if p.GetEntity() != -1 {
		comreq.WriteString(" AND entity.entity_id = :entity")
	}
	if p.GetStorelocation() != -1 {
		comreq.WriteString(" AND storelocation.storelocation_id = :storelocation")
	}
	if p.GetName() != -1 {
		comreq.WriteString(" AND name.name_id = :name")
	}

	// post select request
	postsreq.WriteString(" GROUP BY p.product_id")
	postsreq.WriteString(" ORDER BY " + p.GetOrderBy() + " " + p.GetOrder())

	// limit
	if p.GetLimit() != constants.MaxUint64 {
		postsreq.WriteString(" LIMIT :limit OFFSET :offset")
	}

	// building count and select statements
	if cnstmt, err = db.PrepareNamed(precreq.String() + comreq.String()); err != nil {
		return nil, 0, err
	}
	if snstmt, err = db.PrepareNamed(presreq.String() + comreq.String() + postsreq.String()); err != nil {
		return nil, 0, err
	}

	// building argument map
	m := map[string]interface{}{
		"search":        p.GetSearch(),
		"personid":      p.GetLoggedPersonID(),
		"order":         p.GetOrder(),
		"limit":         p.GetLimit(),
		"offset":        p.GetOffset(),
		"entity":        p.GetEntity(),
		"product":       p.GetProduct(),
		"storelocation": p.GetStorelocation(),
		"name":          p.GetName(),
	}

	// select
	if err = snstmt.Select(&products, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	//
	// getting symbols
	//
	for i, p := range products {
		// note: do not modify p but products[i] instead
		req.Reset()
		req.WriteString("SELECT symbol_id, symbol_label, symbol_image FROM symbol")
		req.WriteString(" JOIN productsymbols ON productsymbols.productsymbols_symbol_id = symbol.symbol_id")
		req.WriteString(" JOIN product ON productsymbols.productsymbols_product_id = product.product_id")
		req.WriteString(" WHERE product.product_id = ?")

		if err = db.Select(&products[i].Symbols, req.String(), p.ProductID); err != nil {
			return nil, 0, err
		}
	}

	//
	// getting synonyms
	//
	for i, p := range products {
		// note: do not modify p but products[i] instead
		req.Reset()
		req.WriteString("SELECT name_id, name_label FROM name")
		req.WriteString(" JOIN productsynonyms ON productsynonyms.productsynonyms_name_id = name.name_id")
		req.WriteString(" JOIN product ON productsynonyms.productsynonyms_product_id = product.product_id")
		req.WriteString(" WHERE product.product_id = ?")

		if err = db.Select(&products[i].Synonyms, req.String(), p.ProductID); err != nil {
			return nil, 0, err
		}
	}

	//
	// getting hazard statements
	//
	for i, p := range products {
		// note: do not modify p but products[i] instead
		req.Reset()
		req.WriteString("SELECT hazardstatement_id, hazardstatement_label, hazardstatement_reference FROM hazardstatement")
		req.WriteString(" JOIN producthazardstatements ON producthazardstatements.producthazardstatements_hazardstatement_id = hazardstatement.hazardstatement_id")
		req.WriteString(" JOIN product ON producthazardstatements.producthazardstatements_product_id = product.product_id")
		req.WriteString(" WHERE product.product_id = ?")

		if err = db.Select(&products[i].HazardStatements, req.String(), p.ProductID); err != nil {
			return nil, 0, err
		}
	}

	//
	// getting precautionary statements
	//
	for i, p := range products {
		// note: do not modify p but products[i] instead
		req.Reset()
		req.WriteString("SELECT precautionarystatement_id, precautionarystatement_label, precautionarystatement_reference FROM precautionarystatement")
		req.WriteString(" JOIN productprecautionarystatements ON productprecautionarystatements.productprecautionarystatements_precautionarystatement_id = precautionarystatement.precautionarystatement_id")
		req.WriteString(" JOIN product ON productprecautionarystatements.productprecautionarystatements_product_id = product.product_id")
		req.WriteString(" WHERE product.product_id = ?")

		if err = db.Select(&products[i].PrecautionaryStatements, req.String(), p.ProductID); err != nil {
			return nil, 0, err
		}
	}

	return products, count, nil
}

func (db *SQLiteDataStore) GetProduct(id int) (Product, error) {
	var (
		product Product
		sqlr    string
		err     error
	)

	sqlr = `SELECT product.product_id, 
	product.product_specificity, 
	product_msds,
	product_restricted,
	product_radioactive,
	product_linearformula,
	product_threedformula,
	product_disposalcomment,
	product_remark,
	empiricalformula.empiricalformula_id AS "empiricalformula.empiricalformula_id",
	empiricalformula.empiricalformula_label AS "empiricalformula.empiricalformula_label",
	physicalstate.physicalstate_id AS "physicalstate.physicalstate_id",
	physicalstate.physicalstate_label AS "physicalstate.physicalstate_label",
	signalword.signalword_id AS "signalword.signalword_id",
	signalword.signalword_label AS "signalword.signalword_label",
	classofcompound.classofcompound_id AS "classofcompound.classofcompound_id",
	classofcompound.classofcompound_label AS "classofcompound.classofcompound_label",
	person.person_id AS "person.person_id",
	person.person_email AS "person.person_email",
	name.name_id AS "name.name_id",
	name.name_label AS "name.name_label",
	cenumber.cenumber_id AS "cenumber.cenumber_id",
	cenumber.cenumber_label AS "cenumber.cenumber_label",
	casnumber.casnumber_id AS "casnumber.casnumber_id",
	casnumber.casnumber_label AS "casnumber.casnumber_label"
	FROM product
	JOIN name ON product.name = name.name_id
	JOIN casnumber ON product.casnumber = casnumber.casnumber_id
	LEFT JOIN cenumber ON product.cenumber = cenumber.cenumber_id
	JOIN person ON product.person = person.person_id
	JOIN empiricalformula ON product.empiricalformula = empiricalformula.empiricalformula_id
	LEFT JOIN physicalstate ON product.physicalstate = physicalstate.physicalstate_id
	LEFT JOIN signalword ON product.signalword = signalword.signalword_id
	LEFT JOIN classofcompound ON product.classofcompound = classofcompound.classofcompound_id
	WHERE product_id = ?`
	if err = db.Get(&product, sqlr, id); err != nil {
		return Product{}, err
	}

	//
	// getting symbols
	//
	sqlr = `SELECT symbol_id, symbol_label, symbol_image FROM symbol
	JOIN productsymbols ON productsymbols.productsymbols_symbol_id = symbol.symbol_id
	JOIN product ON productsymbols.productsymbols_product_id = product.product_id
	WHERE product.product_id = ?`
	if err = db.Select(&product.Symbols, sqlr, product.ProductID); err != nil {
		return product, err
	}

	//
	// getting synonyms
	//
	sqlr = `SELECT name_id, name_label FROM name
	JOIN productsynonyms ON productsynonyms.productsynonyms_name_id = name.name_id
	JOIN product ON productsynonyms.productsynonyms_product_id = product.product_id
	WHERE product.product_id = ?`
	if err = db.Select(&product.Synonyms, sqlr, product.ProductID); err != nil {
		return product, err
	}

	//
	// getting hazard statements
	//
	sqlr = `SELECT hazardstatement_id, hazardstatement_label, hazardstatement_reference FROM hazardstatement
	JOIN producthazardstatements ON producthazardstatements.producthazardstatements_hazardstatement_id = hazardstatement.hazardstatement_id
	JOIN product ON producthazardstatements.producthazardstatements_product_id = product.product_id
	WHERE product.product_id = ?`
	if err = db.Select(&product.HazardStatements, sqlr, product.ProductID); err != nil {
		return product, err
	}

	//
	// getting precautionary statements
	//
	sqlr = `SELECT precautionarystatement_id, precautionarystatement_label, precautionarystatement_reference FROM precautionarystatement
	JOIN productprecautionarystatements ON productprecautionarystatements.productprecautionarystatements_precautionarystatement_id = precautionarystatement.precautionarystatement_id
	JOIN product ON productprecautionarystatements.productprecautionarystatements_product_id = product.product_id
	WHERE product.product_id = ?`
	if err = db.Select(&product.PrecautionaryStatements, sqlr, product.ProductID); err != nil {
		return product, err
	}

	log.WithFields(log.Fields{"ID": id, "product": product}).Debug("GetProduct")
	return product, nil
}

func (db *SQLiteDataStore) DeleteProduct(id int) error {
	var (
		sqlr string
		err  error
	)
	// deleting symbols
	sqlr = `DELETE FROM productsymbols WHERE productsymbols.productsymbols_product_id = (?)`
	if _, err = db.Exec(sqlr, id); err != nil {
		return err
	}

	// deleting synonyms
	sqlr = `DELETE FROM productsynonyms WHERE productsynonyms.productsynonyms_product_id = (?)`
	if _, err = db.Exec(sqlr, id); err != nil {
		return err
	}

	// deleting hazard statements
	sqlr = `DELETE FROM producthazardstatements WHERE producthazardstatements.producthazardstatements_product_id = (?)`
	if _, err = db.Exec(sqlr, id); err != nil {
		return err
	}

	// deleting precautionary statements
	sqlr = `DELETE FROM productprecautionarystatements WHERE productprecautionarystatements.productprecautionarystatements_product_id = (?)`
	if _, err = db.Exec(sqlr, id); err != nil {
		return err
	}

	// deleting product
	sqlr = `DELETE FROM product WHERE product_id = ?`
	if _, err = db.Exec(sqlr, id); err != nil {
		return err
	}
	return nil
}

func (db *SQLiteDataStore) CreateProduct(p Product) (error, int) {
	var (
		lastid   int64
		tx       *sql.Tx
		sqlr     string
		res      sql.Result
		sqla     []interface{}
		ibuilder sq.InsertBuilder
		err      error
	)

	// beginning transaction
	if tx, err = db.Begin(); err != nil {
		return err, 0
	}

	// if CasNumberID = -1 then it is a new cas
	if p.CasNumber.CasNumberID == -1 {
		sqlr = `INSERT INTO casnumber (casnumber_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.CasNumberLabel); err != nil {
			tx.Rollback()
			return err, 0
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err, 0
		}
		// updating the product CasNumberID (CasNumberLabel already set)
		p.CasNumber.CasNumberID = int(lastid)
	}
	// if CeNumberID = -1 then it is a new ce
	if v, err := p.CeNumber.CeNumberID.Value(); p.CeNumber.CeNumberID.Valid && err == nil && v.(int64) == -1 {
		sqlr = `INSERT INTO cenumber (cenumber_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.CeNumberLabel); err != nil {
			tx.Rollback()
			return err, 0
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err, 0
		}
		// updating the product CeNumberID (CeNumberLabel already set)
		p.CeNumber.CeNumberID = sql.NullInt64{Int64: lastid}
	}
	if err != nil {
		log.Error("cenumber error - " + err.Error())
		tx.Rollback()
		return err, 0
	}
	// if NameID = -1 then it is a new name
	if p.Name.NameID == -1 {
		sqlr = `INSERT INTO name (name_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, strings.ToUpper(p.NameLabel)); err != nil {
			tx.Rollback()
			return err, 0
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err, 0
		}
		// updating the product NameID (NameLabel already set)
		p.Name.NameID = int(lastid)
	}
	for i, syn := range p.Synonyms {
		if syn.NameID == -1 {
			sqlr = `INSERT INTO name (name_label) VALUES (?)`
			if res, err = tx.Exec(sqlr, strings.ToUpper(syn.NameLabel)); err != nil {
				tx.Rollback()
				return err, 0
			}
			// getting the last inserted id
			if lastid, err = res.LastInsertId(); err != nil {
				tx.Rollback()
				return err, 0
			}
			p.Synonyms[i].NameID = int(lastid)
		}
	}
	// if EmpiricalFormulaID = -1 then it is a new empirical formula
	if p.EmpiricalFormula.EmpiricalFormulaID == -1 {
		sqlr = `INSERT INTO empiricalformula (empiricalformula_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.EmpiricalFormulaLabel); err != nil {
			tx.Rollback()
			return err, 0
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err, 0
		}
		// updating the product EmpiricalFormulaID (EmpiricalFormulaLabel already set)
		p.EmpiricalFormula.EmpiricalFormulaID = int(lastid)
	}
	// if ClassOfCompoundID = -1 then it is a new class of compound
	if v, err := p.ClassOfCompound.ClassOfCompoundID.Value(); p.ClassOfCompound.ClassOfCompoundID.Valid && err == nil && v.(int64) == -1 {
		sqlr = `INSERT INTO classofcompound (classofcompound_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.ClassOfCompoundLabel); err != nil {
			tx.Rollback()
			return err, 0
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err, 0
		}
		// updating the product ClassOfCompoundID (ClassOfCompoundLabel already set)
		p.ClassOfCompound.ClassOfCompoundID = sql.NullInt64{Int64: lastid}
	}
	if err != nil {
		log.Error("classofcompound error - " + err.Error())
		tx.Rollback()
		return err, 0
	}

	// finally updating the product
	s := make(map[string]interface{})
	if p.ProductSpecificity.Valid {
		s["product_specificity"] = p.ProductSpecificity.String
	}
	if p.ProductMSDS.Valid {
		s["product_msds"] = p.ProductMSDS.String
	}
	if p.ProductRestricted.Valid {
		s["product_restricted"] = p.ProductRestricted.Bool
	}
	if p.ProductRadioactive.Valid {
		s["product_radioactive"] = p.ProductRadioactive.Bool
	}
	if p.ProductLinearFormula.Valid {
		s["product_linearformula"] = p.ProductLinearFormula.String
	}
	if p.ProductThreeDFormula.Valid {
		s["product_threedformula"] = p.ProductThreeDFormula.String
	}
	if p.ProductDisposalComment.Valid {
		s["product_disposalcomment"] = p.ProductDisposalComment.String
	}
	if p.ProductRemark.Valid {
		s["product_remark"] = p.ProductRemark.String
	}
	if p.PhysicalStateID.Valid {
		s["physicalstate"] = int(p.PhysicalStateID.Int64)
	}
	if p.SignalWordID.Valid {
		s["signalword"] = int(p.SignalWordID.Int64)
	}
	if p.ClassOfCompoundID.Valid {
		s["classofcompound"] = int(p.ClassOfCompoundID.Int64)
	}
	if p.CeNumberID.Valid {
		s["cenumber"] = int(p.CeNumberID.Int64)
	}
	s["casnumber"] = p.CasNumberID
	s["name"] = p.NameID
	s["empiricalformula"] = p.EmpiricalFormulaID
	s["person"] = p.PersonID

	// building column names/values
	col := make([]string, 0, len(s))
	val := make([]interface{}, 0, len(s))
	for k, v := range s {
		col = append(col, k)
		rt := reflect.TypeOf(v)
		rv := reflect.ValueOf(v)
		switch rt.Kind() {
		case reflect.Int:
			val = append(val, strconv.Itoa(int(rv.Int())))
		case reflect.String:
			val = append(val, rv.String())
		case reflect.Bool:
			val = append(val, rv.Bool())
		default:
			panic("unknown type:" + rt.String())
		}
	}

	ibuilder = sq.Insert("product").Columns(col...).Values(val...)
	if sqlr, sqla, err = ibuilder.ToSql(); err != nil {
		tx.Rollback()
		return err, 0
	}

	if res, err = tx.Exec(sqlr, sqla...); err != nil {
		log.Error("product error - " + err.Error())
		log.Error("sql:" + sqlr)
		tx.Rollback()
		return err, 0
	}

	// getting the last inserted id
	if lastid, err = res.LastInsertId(); err != nil {
		tx.Rollback()
		return err, 0
	}
	p.ProductID = int(lastid)
	log.WithFields(log.Fields{"p": p}).Debug("CreateProduct")

	// adding symbols
	for _, sym := range p.Symbols {
		sqlr = `INSERT INTO productsymbols (productsymbols_product_id, productsymbols_symbol_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, sym.SymbolID); err != nil {
			log.Error("productsymbols error - " + err.Error())
			tx.Rollback()
			return err, 0
		}
	}
	// adding hazard statements
	for _, hs := range p.HazardStatements {
		sqlr = `INSERT INTO producthazardstatements (producthazardstatements_product_id, producthazardstatements_hazardstatement_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, hs.HazardStatementID); err != nil {
			log.Error("producthazardstatements error - " + err.Error())
			tx.Rollback()
			return err, 0
		}
	}
	// adding precautionary statements
	for _, ps := range p.PrecautionaryStatements {
		sqlr = `INSERT INTO productprecautionarystatements (productprecautionarystatements_product_id, productprecautionarystatements_precautionarystatement_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, ps.PrecautionaryStatementID); err != nil {
			log.Error("productprecautionarystatements error - " + err.Error())
			tx.Rollback()
			return err, 0
		}
	}
	// adding synonyms
	for _, syn := range p.Synonyms {
		sqlr = `INSERT INTO productsynonyms (productsynonyms_product_id, productsynonyms_name_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, syn.NameID); err != nil {
			tx.Rollback()
			return err, 0
		}
	}

	// committing changes
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err, 0
	}

	return nil, p.ProductID
}

func (db *SQLiteDataStore) UpdateProduct(p Product) error {
	var (
		lastid   int64
		tx       *sql.Tx
		sqlr     string
		res      sql.Result
		sqla     []interface{}
		ubuilder sq.UpdateBuilder
		err      error
	)

	// beginning transaction
	if tx, err = db.Begin(); err != nil {
		return err
	}

	// if CasNumberID = -1 then it is a new cas
	if p.CasNumber.CasNumberID == -1 {
		sqlr = `INSERT INTO casnumber (casnumber_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.CasNumberLabel); err != nil {
			tx.Rollback()
			return err
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err
		}
		// updating the product CasNumberID (CasNumberLabel already set)
		p.CasNumber.CasNumberID = int(lastid)
	}
	// if CeNumberID = -1 then it is a new ce
	if v, err := p.CeNumber.CeNumberID.Value(); p.CeNumber.CeNumberID.Valid && err == nil && v.(int64) == -1 {
		sqlr = `INSERT INTO cenumber (cenumber_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.CeNumberLabel); err != nil {
			tx.Rollback()
			return err
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err
		}
		// updating the product CeNumberID (CeNumberLabel already set)
		p.CeNumber.CeNumberID = sql.NullInt64{Int64: lastid}
	}
	if err != nil {
		log.Error("cenumber error - " + err.Error())
		tx.Rollback()
		return err
	}
	// if NameID = -1 then it is a new name
	if p.Name.NameID == -1 {
		sqlr = `INSERT INTO name (name_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, strings.ToUpper(p.NameLabel)); err != nil {
			tx.Rollback()
			return err
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err
		}
		// updating the product NameID (NameLabel already set)
		p.Name.NameID = int(lastid)
	}
	for i, syn := range p.Synonyms {
		if syn.NameID == -1 {
			sqlr = `INSERT INTO name (name_label) VALUES (?)`
			if res, err = tx.Exec(sqlr, strings.ToUpper(syn.NameLabel)); err != nil {
				tx.Rollback()
				return err
			}
			// getting the last inserted id
			if lastid, err = res.LastInsertId(); err != nil {
				tx.Rollback()
				return err
			}
			p.Synonyms[i].NameID = int(lastid)
		}
	}
	// if EmpiricalFormulaID = -1 then it is a new empirical formula
	if p.EmpiricalFormula.EmpiricalFormulaID == -1 {
		sqlr = `INSERT INTO empiricalformula (empiricalformula_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.EmpiricalFormulaLabel); err != nil {
			tx.Rollback()
			return err
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err
		}
		// updating the product EmpiricalFormulaID (EmpiricalFormulaLabel already set)
		p.EmpiricalFormula.EmpiricalFormulaID = int(lastid)
	}
	// if ClassOfCompoundID = -1 then it is a new class of compound
	if v, err := p.ClassOfCompound.ClassOfCompoundID.Value(); p.ClassOfCompound.ClassOfCompoundID.Valid && err == nil && v.(int64) == -1 {
		sqlr = `INSERT INTO classofcompound (classofcompound_label) VALUES (?)`
		if res, err = tx.Exec(sqlr, p.ClassOfCompoundLabel); err != nil {
			tx.Rollback()
			return err
		}
		// getting the last inserted id
		if lastid, err = res.LastInsertId(); err != nil {
			tx.Rollback()
			return err
		}
		// updating the product ClassOfCompoundID (ClassOfCompoundLabel already set)
		p.ClassOfCompound.ClassOfCompoundID = sql.NullInt64{Int64: lastid}
	}
	if err != nil {
		log.Error("classofcompound error - " + err.Error())
		tx.Rollback()
		return err
	}

	// finally updating the product
	s := make(map[string]interface{})
	if p.ProductSpecificity.Valid {
		s["product_specificity"] = p.ProductSpecificity.String
	}
	if p.ProductMSDS.Valid {
		s["product_msds"] = p.ProductMSDS.String
	}
	if p.ProductRestricted.Valid {
		s["product_restricted"] = p.ProductRestricted.Bool
	}
	if p.ProductRadioactive.Valid {
		s["product_radioactive"] = p.ProductRadioactive.Bool
	}
	if p.ProductLinearFormula.Valid {
		s["product_linearformula"] = p.ProductLinearFormula.String
	}
	if p.ProductThreeDFormula.Valid {
		s["product_threedformula"] = p.ProductThreeDFormula.String
	}
	if p.ProductDisposalComment.Valid {
		s["product_disposalcomment"] = p.ProductDisposalComment.String
	}
	if p.ProductRemark.Valid {
		s["product_remark"] = p.ProductRemark.String
	}
	if p.PhysicalStateID.Valid {
		s["physicalstate"] = int(p.PhysicalStateID.Int64)
	}
	if p.SignalWordID.Valid {
		s["signalword"] = int(p.SignalWordID.Int64)
	}
	if p.ClassOfCompoundID.Valid {
		s["classofcompound"] = int(p.ClassOfCompoundID.Int64)
	}
	if p.CeNumberID.Valid {
		s["cenumber"] = int(p.CeNumberID.Int64)
	}
	s["casnumber"] = p.CasNumberID
	s["name"] = p.NameID
	s["empiricalformula"] = p.EmpiricalFormulaID
	s["person"] = p.PersonID

	ubuilder = sq.Update("product").
		SetMap(s).
		Where(sq.Eq{"product_id": p.ProductID})
	if sqlr, sqla, err = ubuilder.ToSql(); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(sqlr, sqla...); err != nil {
		tx.Rollback()
		return err
	}

	// deleting symbols
	sqlr = `DELETE FROM productsymbols WHERE productsymbols.productsymbols_product_id = (?)`
	if res, err = tx.Exec(sqlr, p.ProductID); err != nil {
		tx.Rollback()
		return err
	}
	// adding new ones
	for _, sym := range p.Symbols {
		sqlr = `INSERT INTO productsymbols (productsymbols_product_id, productsymbols_symbol_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, sym.SymbolID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// deleting synonyms
	sqlr = `DELETE FROM productsynonyms WHERE productsynonyms.productsynonyms_product_id = (?)`
	if res, err = tx.Exec(sqlr, p.ProductID); err != nil {
		tx.Rollback()
		return err
	}
	// adding new ones
	for _, syn := range p.Synonyms {
		sqlr = `INSERT INTO productsynonyms (productsynonyms_product_id, productsynonyms_name_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, syn.NameID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// deleting hazard statements
	sqlr = `DELETE FROM producthazardstatements WHERE producthazardstatements.producthazardstatements_product_id = (?)`
	if res, err = tx.Exec(sqlr, p.ProductID); err != nil {
		tx.Rollback()
		return err
	}
	// adding new ones
	for _, hs := range p.HazardStatements {
		sqlr = `INSERT INTO producthazardstatements (producthazardstatements_product_id, producthazardstatements_hazardstatement_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, hs.HazardStatementID); err != nil {
			log.Error("producthazardstatements error - " + err.Error())
			tx.Rollback()
			return err
		}
	}

	// deleting precautionary statements
	sqlr = `DELETE FROM productprecautionarystatements WHERE productprecautionarystatements.productprecautionarystatements_product_id = (?)`
	if res, err = tx.Exec(sqlr, p.ProductID); err != nil {
		tx.Rollback()
		return err
	}
	// adding new ones
	for _, ps := range p.PrecautionaryStatements {
		sqlr = `INSERT INTO productprecautionarystatements (productprecautionarystatements_product_id, productprecautionarystatements_precautionarystatement_id) VALUES (?,?)`
		if res, err = tx.Exec(sqlr, p.ProductID, ps.PrecautionaryStatementID); err != nil {
			log.Error("productprecautionarystatements error - " + err.Error())
			tx.Rollback()
			return err
		}
	}

	// committing changes
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
