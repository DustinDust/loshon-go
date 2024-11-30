create table
  public.documents (
    id uuid not null default gen_random_uuid (),
    created_at timestamp with time zone null,
    updated_at timestamp with time zone null,
    deleted_at timestamp with time zone null,
    title text null,
    user_id text null,
    is_archived boolean null,
    parent_document_id uuid null,
    content text null,
		md_content text null,
    cover_image text null,
    icon text null,
    is_published boolean null,
    constraint documents_pkey primary key (id),
    constraint fk_documents_child_documents foreign key (parent_document_id) references documents (id)
  ) tablespace pg_default;

create index if not exists idx_documents_deleted_at on public.documents using btree (deleted_at) tablespace pg_default;

create index if not exists idx_documents_user_id on public.documents using btree (user_id) tablespace pg_default;

create index if not exists idx_documents_id on public.documents using btree (id) tablespace pg_default;

create index if not exists idx_documents_parent_document on public.documents using btree (parent_document_id) tablespace pg_default;

create index if not exists idx_documents_parent_document_id on public.documents using btree (parent_document_id) tablespace pg_default;