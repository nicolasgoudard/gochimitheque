package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3" // register sqlite3 driver
	log "github.com/sirupsen/logrus"
	"github.com/tbellembois/gochimitheque/constants"
	"github.com/tbellembois/gochimitheque/helpers"
)

type stockMapKey struct {
	sid int // store location id
	uid int // init id
}

type stockMapValue struct {
	t int // total
	c int // current
}

type StockMap map[stockMapKey]stockMapValue

func (db *SQLiteDataStore) ComputeStockStorelocation(p Product, s StoreLocation, u Unit, m *StockMap) int {

	var (
		i   int
		err error
	)

	sqlr := `SELECT storage.storage_quantity FROM storage
	WHERE storage.storelocation = ? AND
	storage.product = ? AND
	storage.unit = ?`

	fmt.Println(s.StoreLocationID.Int64)
	fmt.Println(p.ProductID)
	fmt.Println(u.UnitID.Int64)

	if err = db.Get(&i, sqlr, s.StoreLocationID.Int64, p.ProductID, u.UnitID.Int64); err != nil {
		fmt.Println(err.Error())
		return 0
	}

	log.Info(i)

	return i
}

// GetEntities returns the entities matching the search criteria
// order, offset and limit are passed to the sql request
func (db *SQLiteDataStore) GetEntities(p helpers.DbselectparamEntity) ([]Entity, int, error) {
	var (
		entities                                []Entity
		count                                   int
		req, precreq, presreq, comreq, postsreq strings.Builder
		cnstmt                                  *sqlx.NamedStmt
		snstmt                                  *sqlx.NamedStmt
		err                                     error
	)
	log.WithFields(log.Fields{"p": p}).Debug("GetEntities")

	precreq.WriteString(" SELECT count(DISTINCT e.entity_id)")
	presreq.WriteString(" SELECT e.entity_id, e.entity_name, e.entity_description")
	comreq.WriteString(" FROM entity AS e, person as p")
	// filter by permissions
	comreq.WriteString(` JOIN permission AS perm ON
	(perm.person = :personid and perm.permission_item_name = "all" and perm.permission_perm_name = "all" and perm.permission_entity_id = e.entity_id) OR
	(perm.person = :personid and perm.permission_item_name = "all" and perm.permission_perm_name = "all" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "all" and perm.permission_perm_name = "r" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "entities" and perm.permission_perm_name = "all" and perm.permission_entity_id = e.entity_id) OR
	(perm.person = :personid and perm.permission_item_name = "entities" and perm.permission_perm_name = "all" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "entities" and perm.permission_perm_name = "r" and perm.permission_entity_id = -1) OR
	(perm.person = :personid and perm.permission_item_name = "entities" and perm.permission_perm_name = "r" and perm.permission_entity_id = e.entity_id)
	`)
	comreq.WriteString(" WHERE e.entity_name LIKE :search")
	postsreq.WriteString(" GROUP BY e.entity_id")
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
		"search":   p.GetSearch(),
		"personid": p.GetLoggedPersonID(),
		"order":    p.GetOrder(),
		"limit":    p.GetLimit(),
		"offset":   p.GetOffset(),
	}

	// select
	if err = snstmt.Select(&entities, m); err != nil {
		return nil, 0, err
	}
	// count
	if err = cnstmt.Get(&count, m); err != nil {
		return nil, 0, err
	}

	//
	// getting managers
	//
	for i, e := range entities {
		// note: do not modify e but entities[i] instead
		req.Reset()
		req.WriteString("SELECT person_id, person_email FROM person")
		req.WriteString(" JOIN entitypeople ON entitypeople.entitypeople_person_id = person.person_id")
		req.WriteString(" JOIN entity ON entitypeople.entitypeople_entity_id = entity.entity_id")
		req.WriteString(" WHERE entity.entity_id = ?")

		if err = db.Select(&entities[i].Managers, req.String(), e.EntityID); err != nil {
			return nil, 0, err
		}
	}

	log.WithFields(log.Fields{"entities": entities, "count": count}).Debug("GetEntities")
	return entities, count, nil
}

// GetEntity returns the entity with id "id"
func (db *SQLiteDataStore) GetEntity(id int) (Entity, error) {
	var (
		entity Entity
		sqlr   string
		err    error
	)
	log.WithFields(log.Fields{"id": id}).Debug("GetEntity")

	sqlr = `SELECT e.entity_id, e.entity_name, e.entity_description
	FROM entity AS e
	WHERE e.entity_id = ?`
	if err = db.Get(&entity, sqlr, id); err != nil {
		return Entity{}, err
	}
	log.WithFields(log.Fields{"ID": id, "entity": entity}).Debug("GetEntity")
	return entity, nil
}

// GetEntityPeople returns the entity (with id "id") managers
func (db *SQLiteDataStore) GetEntityPeople(id int) ([]Person, error) {
	var (
		people []Person
		sqlr   string
		err    error
	)

	sqlr = `SELECT p.person_id, p.person_email
	FROM person AS p, entitypeople
	WHERE entitypeople.entitypeople_person_id == p.person_id AND entitypeople.entitypeople_entity_id = ?`
	if err = db.Select(&people, sqlr, id); err != nil {
		return []Person{}, err
	}
	log.WithFields(log.Fields{"ID": id, "people": people}).Debug("GetEntityPeople")
	return people, nil
}

// DeleteEntity deletes the entity with id "id"
func (db *SQLiteDataStore) DeleteEntity(id int) error {
	var (
		sqlr string
		err  error
	)
	sqlr = `DELETE FROM entity 
	WHERE entity_id = ?`
	if _, err = db.Exec(sqlr, id); err != nil {
		return err
	}
	return nil
}

// CreateEntity creates the given entity
func (db *SQLiteDataStore) CreateEntity(e Entity) (error, int) {
	var (
		sqlr   string
		res    sql.Result
		lastid int64
		err    error
	)
	// FIXME: use a transaction here
	sqlr = `INSERT INTO entity(entity_name, entity_description) VALUES (?, ?)`
	if res, err = db.Exec(sqlr, e.EntityName, e.EntityDescription); err != nil {
		return err, 0
	}

	// getting the last inserted id
	if lastid, err = res.LastInsertId(); err != nil {
		return err, 0
	}
	e.EntityID = int(lastid)

	// adding the new managers
	for _, m := range e.Managers {
		sqlr = `INSERT INTO entitypeople (entitypeople_entity_id, entitypeople_person_id) values (?, ?)`
		if _, err = db.Exec(sqlr, e.EntityID, m.PersonID); err != nil {
			return err, 0
		}

		// setting the manager in the entity
		sqlr = `INSERT OR IGNORE INTO personentities(personentities_person_id, personentities_entity_id) 
			VALUES (?, ?)`
		if _, err = db.Exec(sqlr, m.PersonID, e.EntityID); err != nil {
			return err, 0
		}

		// setting the manager permissions in the entity
		// 1. lazily deleting former permissions
		sqlr = `DELETE FROM permission 
			WHERE person = ? and permission_entity_id = ?`
		if _, err = db.Exec(sqlr, m.PersonID, e.EntityID); err != nil {
			return err, 0
		}
		// 2. inserting manager permissions
		sqlr = `INSERT INTO permission(person, permission_perm_name, permission_item_name, permission_entity_id) 
			VALUES (?, ?, ?, ?)`
		if _, err = db.Exec(sqlr, m.PersonID, "all", "all", e.EntityID); err != nil {
			return err, 0
		}
	}

	return nil, e.EntityID
}

// UpdateEntity updates the given entity
func (db *SQLiteDataStore) UpdateEntity(e Entity) error {
	var (
		sqlr     string
		sqla     []interface{}
		sbuilder sq.DeleteBuilder
		err      error
	)
	log.WithFields(log.Fields{"e": e}).Debug("UpdateEntity")

	// updating the entity
	// FIXME: use a transaction here
	sqlr = `UPDATE entity SET entity_name = ?, entity_description = ?
	WHERE entity_id = ?`
	if _, err = db.Exec(sqlr, e.EntityName, e.EntityDescription, e.EntityID); err != nil {
		return err
	}

	if len(e.Managers) != 0 {
		// removing former managers
		notin := sq.Or{}
		// ex: AND (entitypeople_person_id <> ? OR entitypeople_person_id <> ?)
		for _, m := range e.Managers {
			notin = append(notin, sq.NotEq{"entitypeople_person_id": m.PersonID})
		}
		// ex: DELETE FROM entitypeople WHERE (entitypeople_entity_id = ? AND (entitypeople_person_id <> ? OR entitypeople_person_id <> ?)
		sbuilder = sq.Delete(`entitypeople`).Where(
			sq.And{
				sq.Eq{`entitypeople_entity_id`: e.EntityID},
				notin})
	} else {
		sbuilder = sq.Delete(`entitypeople`).Where(
			sq.Eq{`entitypeople_entity_id`: e.EntityID})
	}
	sqlr, sqla, err = sbuilder.ToSql()
	if err != nil {
		return err
	}
	if _, err = db.Exec(sqlr, sqla...); err != nil {
		return err
	}

	// TODO: removing former managers permissions

	// adding the new ones
	for _, m := range e.Managers {
		// adding the manager
		sqlr = `INSERT OR IGNORE INTO entitypeople (entitypeople_entity_id, entitypeople_person_id) VALUES (?, ?)`
		if _, err = db.Exec(sqlr, e.EntityID, m.PersonID); err != nil {
			return err
		}

		for _, man := range e.Managers {
			// setting the manager in the entity
			sqlr = `INSERT OR IGNORE INTO personentities(personentities_person_id, personentities_entity_id) 
			VALUES (?, ?)`
			if _, err = db.Exec(sqlr, man.PersonID, e.EntityID); err != nil {
				return err
			}

			// setting the manager permissions in the entity
			// 1. lazily deleting former permissions
			sqlr = `DELETE FROM permission 
			WHERE person = ? and permission_entity_id = ?`
			if _, err = db.Exec(sqlr, man.PersonID, e.EntityID); err != nil {
				return err
			}
			// 2. inserting manager permissions
			sqlr = `INSERT INTO permission(person, permission_perm_name, permission_item_name, permission_entity_id) 
			VALUES (?, ?, ?, ?)`
			if _, err = db.Exec(sqlr, man.PersonID, "all", "all", e.EntityID); err != nil {
				return err
			}

		}
	}

	return nil
}

// IsEntityEmpty returns true is the entity is empty
func (db *SQLiteDataStore) IsEntityEmpty(id int) (bool, error) {
	var (
		res   bool
		count int
		sqlr  string
		err   error
	)

	sqlr = "SELECT count(*) from personentities WHERE personentities.personentities_entity_id = ?"
	if err = db.Get(&count, sqlr, id); err != nil {
		return false, err
	}
	log.WithFields(log.Fields{"id": id, "count": count}).Debug("IsEntityEmpty")
	if count == 0 {
		res = true
	} else {
		res = false
	}
	return res, nil
}
