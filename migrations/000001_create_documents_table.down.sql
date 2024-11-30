drop index if exists idx_documents_deleted_at;

drop index if exists idx_documents_user_id;

drop index if exists idx_documents_id;

drop index if exists idx_documents_parent_document;

drop index if exists idx_documents_parent_document_id;

drop table if exists public.documents cascade;