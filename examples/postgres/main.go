package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/sishui/bake/examples/postgres/model"
)

const dsn = "postgres://postgres:123456@127.0.0.1:5432/postgres?sslmode=disable"

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
	var user model.User
	err = db.NewSelect().
		Model(&user).
		Where(model.UserIDEq, 1).
		Relation(model.UserRelationPosts).
		Scan(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(user.ID, user.Name)
	for _, post := range user.Posts {
		fmt.Println(post.ID, post.Title)
	}

	var posts []*model.Post
	err = db.NewSelect().
		Model(&posts).
		Relation(model.PostRelationUser).
		Scan(ctx)
	if err != nil {
		panic(err)
	}
	for _, post := range posts {
		fmt.Println(post.ID, post.Title, post.User.ID, post.User.Name)
	}
	date, err := selectDate(ctx, db, model.UserTableName, model.UserCreatedAtDateExpr)
	if err != nil {
		panic(err)
	}
	fmt.Println(date)
	years, err := selectYMD(ctx, db, model.UserTableName, model.UserCreatedAtYearExpr)
	if err != nil {
		panic(err)
	}
	fmt.Println(years)
	months, err := selectYMD(ctx, db, model.UserTableName, model.UserCreatedAtMonthExpr)
	if err != nil {
		panic(err)
	}
	fmt.Println(months)
	days, err := selectYMD(ctx, db, model.UserTableName, model.UserCreatedAtDayExpr)
	if err != nil {
		panic(err)
	}
	fmt.Println(days)
	hours, err := selectYMD(ctx, db, model.UserTableName, model.UserCreatedAtHourExpr)
	if err != nil {
		panic(err)
	}
	fmt.Println(hours)
	fmt.Println("done")
}

func newDB(ctx context.Context) *bun.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	client := bun.NewDB(db, pgdialect.New())

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxIdleTime(time.Minute)
	db.SetConnMaxLifetime(time.Minute * 3)
	client = client.WithQueryHook(bundebug.NewQueryHook(bundebug.WithEnabled(true), bundebug.WithVerbose(true)))
	return client
}

func selectYMD(ctx context.Context, db *bun.DB, table string, expr string) ([]int, error) {
	var n []int
	err := db.NewSelect().
		Table(table).
		ColumnExpr(expr).
		Scan(ctx, &n)
	if err != nil {
		return nil, err
	}
	return n, err
}

func selectDate(ctx context.Context, db *bun.DB, table string, expr string) ([]time.Time, error) {
	var n []time.Time
	err := db.NewSelect().
		Table(table).
		ColumnExpr(expr).
		Scan(ctx, &n)
	if err != nil {
		return nil, err
	}
	return n, err
}
