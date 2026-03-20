package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/sishui/bake/examples/mysql/model"
)

const dsn = "root:123456@tcp(127.0.0.1:3306)/bake?charset=utf8mb4&parseTime=True&loc=Local"

func main() {
	ctx := context.Background()
	db := newDB(ctx)
	defer db.Close()
	var mails []*model.Mail
	err := db.NewSelect().
		Model(&mails).
		Relation(model.MailRelationAttachments).
		Scan(ctx)
	if err != nil {
		panic(err)
	}
	for _, mail := range mails {
		fmt.Println(mail.ID, mail.Subject)
		for _, attachment := range mail.Attachments {
			fmt.Println(attachment.ID, attachment.MailID)
		}
	}
	var data model.TestAllType
	err = db.NewSelect().Model(&data).Where(model.TestAllTypeIDEq, 1).Scan(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", data)
	fmt.Println("done")
}

func newDB(ctx context.Context) *bun.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	client := bun.NewDB(db, mysqldialect.New())

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxIdleTime(time.Minute)
	db.SetConnMaxLifetime(time.Minute * 3)
	client = client.WithQueryHook(bundebug.NewQueryHook(bundebug.WithEnabled(true), bundebug.WithVerbose(true)))
	return client
}
