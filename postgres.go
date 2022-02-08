package main

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"time"
)

type (
	Hit struct {
		ID      uint64 `pg:"id,pk" json:"id"`
		Device  string `pg:"device" json:"device"`
		ShortId uint64 `pg:"short_id" json:"short_id"`
	}
	Shorts struct {
		ID          int64 `pg:"id,pk"`
		OriginalURL string
		Hits        []Hit `pg:"rel:has-many,join_fk:short_id"`
		// Hits        int64
	}
)

func getConnection(ctx context.Context) (con *pg.DB) {
	db := pg.Connect(&pg.Options{
		Addr:     "postgres:5432",
		User:     "postgres",
		Password: "postgres",
		Database: "postgres",
	})
	for {
		err := db.Ping(ctx)
		if err == nil {
			break
		}
		fmt.Println("will retry in 3 seconds")
		time.Sleep(3 * time.Second)
	}
	fmt.Println("Connected to Postgres")
	return db
}

func getHitCount(keyword string, con *pg.DB) (int, error) {
	id, err := Decode(keyword)
	if err != nil {
		return 0, err
	}
	return con.Model(&Hit{}).Where("short_id = ?", id).Count()
}

func getHits(keyword string, con *pg.DB) (interface{}, error) {
	var res []struct {
		Device   string
		HitCount int
	}
	id, err := Decode(keyword)
	err = con.Model(&Hit{}).Where("short_id = ?", id).Column("device").ColumnExpr("count(*) AS hit_count").Group("device").
		OrderExpr("hit_count DESC").Select(&res)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res, nil
}

func createShortUrl(longurl string, con *pg.DB) (string, error) {
	s := &Shorts{
		OriginalURL: longurl,
	}
	_, err := con.Model(s).Insert()
	shortCode := Encode(uint64(s.ID))
	if err != nil {
		return "", err
	}
	return shortCode, nil
}

func registerHit(keyword string, deviceType string, con *pg.DB) (string, error) {
	id, err := Decode(keyword)
	hit := &Hit{
		ShortId: id,
		Device:  deviceType,
	}
	_, err = con.Model(hit).Insert()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return "", nil
}

func getOriginalUrl(shortCode string, con *pg.DB) (string, error) {
	id, err := Decode(shortCode)
	if err != nil {
		return "", err
	}
	s := &Shorts{}
	err = con.Model(s).Where("id = ?", id).Select()
	if err != nil {
		return "", err
	}
	return s.OriginalURL, nil
}

func migrate(db *pg.DB) error {
	models := []interface{}{
		(*Shorts)(nil),
		(*Hit)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false,
			FKConstraints: true,
			IfNotExists:   true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
