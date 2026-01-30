package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment/domain"
	contestPort "github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application/port"
	commonDto "github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
)

type CommentService struct {
	commentRepo port.CommentDatabasePort
	contestRepo contestPort.ContestDatabasePort
}

func NewCommentService(
	commentRepo port.CommentDatabasePort,
	contestRepo contestPort.ContestDatabasePort,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		contestRepo: contestRepo,
	}
}

func (s *CommentService) CreateComment(contestID, userID int64, req *dto.CreateCommentRequest) (*dto.CommentResponse, error) {
	_, err := s.contestRepo.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	comment := domain.NewComment(contestID, userID, req.Content)

	if err := comment.Validate(); err != nil {
		return nil, err
	}

	savedComment, err := s.commentRepo.Save(comment)
	if err != nil {
		return nil, err
	}

	author := &dto.AuthorResponse{
		UserID: userID,
	}

	return dto.ToCommentResponse(savedComment, author), nil
}

func (s *CommentService) GetCommentByID(commentID int64) (*domain.Comment, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) GetCommentsByContestID(
	contestID int64,
	pagination *commonDto.PaginationRequest,
	sort *commonDto.SortRequest,
) (*commonDto.PaginationResponse, error) {
	_, err := s.contestRepo.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	comments, totalCount, err := s.commentRepo.GetByContestID(contestID, pagination, sort)
	if err != nil {
		return nil, err
	}

	commentResponses := dto.ToCommentResponses(comments)

	return commonDto.NewPaginationResponse(
		commentResponses,
		pagination.Page,
		pagination.PageSize,
		totalCount,
	), nil
}

func (s *CommentService) UpdateComment(commentID, userID int64, req *dto.UpdateCommentRequest) (*dto.CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, err
	}

	if !comment.IsOwner(userID) {
		return nil, exception.ErrCommentPermissionDenied
	}

	if err := comment.UpdateContent(req.Content); err != nil {
		return nil, err
	}

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}

	author := &dto.AuthorResponse{
		UserID: userID,
	}

	return dto.ToCommentResponse(comment, author), nil
}

func (s *CommentService) DeleteComment(commentID, userID int64) error {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}

	if !comment.IsOwner(userID) {
		return exception.ErrCommentPermissionDenied
	}

	return s.commentRepo.Delete(commentID)
}
