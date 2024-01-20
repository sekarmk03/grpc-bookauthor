package services

import (
	"context"
	"grpc-bookauthor/cmd/helpers"
	pb "grpc-bookauthor/proto"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BookService struct {
	pb.UnimplementedBookServiceServer
	DB *gorm.DB
}

func (p *BookService) GetBooks(ctx context.Context, pg *pb.Page) (*pb.Books, error) {
	log.Printf("GetBooks invoked")
	var page int64 = 1
	if pg.GetPage() != 0 {
		page = pg.GetPage()
	}

	pagination := pb.Pagination{}
	books := []*pb.Book{}

	sql := p.DB.Table("books AS b").
		Joins("LEFT JOIN authors AS a ON a.id = b.author_id").
		Select("b.id", "b.title", "b.isbn", "a.id as author_id", "a.name as author_name")

	offset, limit := helpers.Pagination(sql, page, &pagination)

	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			err.Error(),
		)
	}

	defer rows.Close()

	for rows.Next() {
		book := pb.Book{}
		author := pb.Author{}

		if err := rows.Scan(&book.Id, &book.Title, &book.Isbn, &author.Id, &author.Name); err != nil {
			log.Fatalf("Failed get data from database %v", err.Error())
		}

		book.Author = &author
		books = append(books, &book)
	}

	response := &pb.Books{
		Pagination: &pagination,
		Data:       books,
	}

	return response, nil
}

func (p *BookService) GetBook(ctx context.Context, id *pb.Id) (*pb.Book, error) {
	log.Printf("GetBook invoked")
	row := p.DB.Table("books AS b").
		Joins("LEFT JOIN authors AS a ON a.id = b.author_id").
		Select("b.id", "b.title", "b.isbn", "a.id as author_id", "a.name as author_name").
		Where("b.id = ?", id.GetId()).
		Row()

	book := pb.Book{}
	author := pb.Author{}

	if err := row.Scan(&book.Id, &book.Title, &book.Isbn, &author.Id, &author.Name); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			err.Error(),
		)
	}

	book.Author = &author

	return &book, nil
}

func (p *BookService) CreateBook(ctx context.Context, b *pb.Book) (*pb.Id, error) {
	log.Printf("CreateBook invoked")
	Response := pb.Id{}

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		author := pb.Author{
			Id:   0,
			Name: b.GetAuthor().GetName(),
		}

		if err := tx.Table("authors").Where("LCASE(name) = ?", author.GetName()).FirstOrCreate(&author).Error; err != nil {
			return err
		}

		book := struct {
			Id        uint64
			Title     string
			Isbn      string
			Author_id uint32
		}{
			Id:        b.GetId(),
			Title:     b.GetTitle(),
			Isbn:      b.GetIsbn(),
			Author_id: author.GetId(),
		}

		if err := tx.Table("books").Create(&book).Error; err != nil {
			return err
		}

		Response.Id = book.Id
		return nil
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			err.Error(),
		)
	}

	return &Response, nil
}

func (p *BookService) UpdateBook(ctx context.Context, b *pb.Book) (*pb.Status, error) {
	log.Printf("UpdateBook invoked")
	Response := pb.Status{}

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		author := pb.Author{
			Id:   0,
			Name: b.GetAuthor().GetName(),
		}

		if err := tx.Table("authors").Where("LCASE(name) = ?", author.GetName()).FirstOrCreate(&author).Error; err != nil {
			return err
		}

		book := struct {
			Id        uint64
			Title     string
			Isbn      string
			Author_id uint32
		}{
			Id:        b.GetId(),
			Title:     b.GetTitle(),
			Isbn:      b.GetIsbn(),
			Author_id: author.GetId(),
		}

		res := tx.Table("books").Where("id = ?", book.Id).Updates(&book)
		if err := res.Error; err != nil {
			return err
		}

		if res.RowsAffected == 0 {
			Response.Status = 0
			return nil
		}

		Response.Status = 1
		return nil
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			err.Error(),
		)
	}

	if Response.Status == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			"Data not found",
		)
	}

	return &Response, nil
}

func (p *BookService) DeleteBook(ctx context.Context, id *pb.Id) (*pb.Status, error) {
	log.Printf("DeleteBook invoked")
	response := pb.Status{}

	res := p.DB.Table("books").Where("id = ?", id.GetId()).Delete(nil)
	if err := res.Error; err != nil {
		return nil, status.Errorf(
			codes.Internal,
			err.Error(),
		)
	}

	if res.RowsAffected == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			"Data not found",
		)
	}

	response.Status = 1

	return &response, nil
}
