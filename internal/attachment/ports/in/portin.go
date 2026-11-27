package attachmentin

type AttachmentPortIn struct {
	UploadAttachmentUsecase        UploadAttachmentUsecaseInt
	StageUploadUsecase             StageUploadUsecaseInt
	DownloadAttachmentUsecase      DownloadAttachmentUsecaseInt
	GetAttachmentUsecase           GetAttachmentUsecaseInt
	ListAttachmentsUsecase         ListAttachmentsUsecaseInt
	DeleteAttachmentUsecase        DeleteAttachmentUsecaseInt
	DeleteAttachmentsForRefUsecase DeleteAttachmentsForRefUsecaseInt
}

func NewAttachmentPortIn(
	uploadAttachmentUsecase UploadAttachmentUsecaseInt,
	stageUploadUsecase StageUploadUsecaseInt,
	downloadAttachmentUsecase DownloadAttachmentUsecaseInt,
	getAttachmentUsecase GetAttachmentUsecaseInt,
	listAttachmentsUsecase ListAttachmentsUsecaseInt,
	deleteAttachmentUsecase DeleteAttachmentUsecaseInt,
	deleteAttachmentsForRefUsecase DeleteAttachmentsForRefUsecaseInt,
) *AttachmentPortIn {
	return &AttachmentPortIn{
		UploadAttachmentUsecase:        uploadAttachmentUsecase,
		StageUploadUsecase:             stageUploadUsecase,
		DownloadAttachmentUsecase:      downloadAttachmentUsecase,
		GetAttachmentUsecase:           getAttachmentUsecase,
		ListAttachmentsUsecase:         listAttachmentsUsecase,
		DeleteAttachmentUsecase:        deleteAttachmentUsecase,
		DeleteAttachmentsForRefUsecase: deleteAttachmentsForRefUsecase,
	}
}
