package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data sources.

func NewImageDataSource() datasource.DataSource {
	return newManagedDataSource(imageDataSourceDescriptor)
}

func NewImagesDataSource() datasource.DataSource {
	return newManagedDataSource(imagesDataSourceDescriptor)
}

func NewVideoDataSource() datasource.DataSource {
	return newManagedDataSource(videoDataSourceDescriptor)
}

func NewVideosDataSource() datasource.DataSource {
	return newManagedDataSource(videosDataSourceDescriptor)
}

func NewProjectDataSource() datasource.DataSource {
	return newManagedDataSource(projectDataSourceDescriptor)
}

func NewProjectsDataSource() datasource.DataSource {
	return newManagedDataSource(projectsDataSourceDescriptor)
}

func NewApiKeysDataSource() datasource.DataSource {
	return newManagedDataSource(apiKeysDataSourceDescriptor)
}

var imageDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "image",
	SchemaFn:    ImageDataSourceSchema,
	Model:       (*ImageModel)(nil),
	ClientField: "Images",
	ReadMethod:  "GetImage",
	PathParams:  []string{"id"},
}

var imagesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "images",
	SchemaFn:    ImagesDataSourceSchema,
	Model:       (*ImagesModel)(nil),
	ClientField: "Images",
	ReadMethod:  "ListImages",
	PathParams:  []string{"project_id"},
}

var videoDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "video",
	SchemaFn:    VideoDataSourceSchema,
	Model:       (*VideoModel)(nil),
	ClientField: "Videos",
	ReadMethod:  "GetVideo",
	PathParams:  []string{"id"},
}

var videosDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "videos",
	SchemaFn:    VideosDataSourceSchema,
	Model:       (*VideosModel)(nil),
	ClientField: "Videos",
	ReadMethod:  "ListVideos",
	PathParams:  []string{"project_id"},
}

var projectDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "project",
	SchemaFn:    ProjectDataSourceSchema,
	Model:       (*ProjectModel)(nil),
	ClientField: "Projects",
	ReadMethod:  "GetProject",
	PathParams:  []string{"org_id", "id"},
}

var projectsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "projects",
	SchemaFn:    ProjectsDataSourceSchema,
	Model:       (*ProjectsModel)(nil),
	ClientField: "Projects",
	ReadMethod:  "ListProjects",
	PathParams:  []string{"org_id"},
}

var apiKeysDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "api_keys",
	SchemaFn:    ApiKeysDataSourceSchema,
	Model:       (*ApiKeysModel)(nil),
	ClientField: "APIKeys",
	ReadMethod:  "ListApiKeys",
	PathParams:  []string{"org_id"},
}

// Resources.

func NewProjectResource() resource.Resource {
	return newManagedResource(projectResourceDescriptor)
}

var projectResourceDescriptor = ResourceDescriptor{
	TypeName:     "project",
	SchemaFn:     ProjectResourceSchema,
	Model:        (*ProjectModel)(nil),
	ClientField:  "Projects",
	CreateMethod: "CreateProject",
	ReadMethod:   "GetProject",
	DeleteMethod: "DeleteProject",
	PathParams:   []string{"org_id", "id"},
}

// ProjectResourceSchema returns a resource schema derived from the data source
// schema. Resource attributes that may be provided during creation are marked
// optional; read-only attributes remain computed.
func ProjectResourceSchema(ctx context.Context) rschema.Schema {
	ds := ProjectDataSourceSchema(ctx)
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"created_at": rschema.StringAttribute{
				Computed:            true,
				Description:         ds.Attributes["created_at"].GetDescription(),
				MarkdownDescription: ds.Attributes["created_at"].GetMarkdownDescription(),
			},
			"custom_domain": rschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"name": rschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"org_id": rschema.StringAttribute{
				Required: true,
			},
			"regions": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"updated_at": rschema.StringAttribute{
				Computed:            true,
				Description:         ds.Attributes["updated_at"].GetDescription(),
				MarkdownDescription: ds.Attributes["updated_at"].GetMarkdownDescription(),
			},
			"video_quality": rschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}
