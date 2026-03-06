package service

import (
	"context"
	"io"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/2-gRPC/internal/database"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/2-gRPC/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CategoryService struct {
	pb.UnimplementedCategoryServiceServer
	CategoryDB database.Category
}

func NewCategoryService(categoryDB database.Category) *CategoryService {
	return &CategoryService{
		CategoryDB: categoryDB,
	}
}

func (c *CategoryService) CreateCategory(ctx context.Context, in *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	category, err := c.CategoryDB.Create(in.Name, in.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CategoryResponse{
		Category: &pb.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		},
	}, nil
}

func (c *CategoryService) CreateCategoryStream(stream grpc.ClientStreamingServer[pb.CreateCategoryRequest, pb.CategoryList]) error {
	var categoriesResponse []*pb.Category

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.CategoryList{
				Categories: categoriesResponse,
			})
		}

		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		category, err := c.CategoryDB.Create(in.Name, in.Description)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		categoryResponse := &pb.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		}

		categoriesResponse = append(categoriesResponse, categoryResponse)
	}
}

func (c *CategoryService) CreateCategoryStreamBidirectional(stream grpc.BidiStreamingServer[pb.CreateCategoryRequest, pb.Category]) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		category, err := c.CategoryDB.Create(in.Name, in.Description)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		categoryResponse := &pb.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		}

		if err := stream.Send(categoryResponse); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
}

func (c *CategoryService) ListCategories(ctx context.Context, in *pb.Blank) (*pb.CategoryList, error) {
	categories, err := c.CategoryDB.FindAll()
	if err != nil {
		return nil, err
	}

	var categoriesResponse []*pb.Category

	for _, category := range categories {
		categoryResponse := &pb.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		}

		categoriesResponse = append(categoriesResponse, categoryResponse)
	}

	return &pb.CategoryList{Categories: categoriesResponse}, nil
}

func (c *CategoryService) GetCategory(ctx context.Context, in *pb.CategoryGetRequest) (*pb.Category, error) {
	category, err := c.CategoryDB.Find(in.Id)
	if err != nil {
		return nil, err
	}

	categoryResponse := &pb.Category{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}

	return categoryResponse, nil
}
