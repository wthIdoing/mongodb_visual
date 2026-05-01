export interface ConnectionInfo {
  status: string
  version: string
  database: string
  server: string
  round_trip_ms: number
}

export interface ConnectionTestRequest {
  host: string
  port: number
  database: string
  username: string
  password: string
  authSource: string
}

export interface CreateDatabaseRequest {
  database: string
  firstCollection: string
}

export interface ConnectionTestResponse {
  status: string
  version: string
  database: string
  server: string
  round_trip_ms: number
}

export interface ListResponse<T> {
  items: T[]
}

export interface PaginationResponse<T> {
  items: T[]
  page: number
  page_size: number
  total: number
}

export interface DocumentEnvelope {
  item: Record<string, unknown>
}

export interface CollectionItem {
  name: string
  document_count?: number
  is_system: boolean
}

export interface IndexItem {
  name: string
  keys: Record<string, unknown>
  unique: boolean
  sparse: boolean
}
