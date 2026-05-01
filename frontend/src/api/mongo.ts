import axios from 'axios'

import { http } from './http'
import { buildMongoSessionHeaders, getMongoSessionConnection } from '../session/mongoConnection'
import type {
  CollectionItem,
  ConnectionInfo,
  ConnectionTestRequest,
  ConnectionTestResponse,
  CreateDatabaseRequest,
  DocumentEnvelope,
  IndexItem,
  ListResponse,
  PaginationResponse,
} from './types'

export function fetchConnectionInfo() {
  return http.get<ConnectionInfo>('/api/v1/meta/connection').then((res) => res.data)
}

export function testConnection(payload: ConnectionTestRequest) {
  return http
    .post<ConnectionTestResponse>('/api/v1/connection/test', {
      host: payload.host,
      port: String(payload.port),
      database: payload.database,
      username: payload.username,
      password: payload.password,
      auth_source: payload.authSource,
    })
    .then((res) => res.data)
}

export function fetchDatabases() {
  return http.get<ListResponse<string>>('/api/v1/databases').then((res) => res.data)
}

export function createDatabase(payload: CreateDatabaseRequest) {
  return http
    .post('/api/v1/databases', {
      database: payload.database,
      first_collection: payload.firstCollection,
    })
    .then((res) => res.data)
}

export function deleteDatabase(database: string) {
  return http.delete(`/api/v1/databases/${encodeURIComponent(database)}`).then((res) => res.data)
}

export function fetchCollections(database: string, params: { includeSystem?: boolean; includeCounts?: boolean } = {}) {
  return http
    .get<ListResponse<CollectionItem>>(`/api/v1/databases/${encodeURIComponent(database)}/collections`, {
      params: {
        include_system: params.includeSystem,
        include_counts: params.includeCounts,
      },
    })
    .then((res) => res.data)
}

export function createCollection(database: string, name: string) {
  return http.post(`/api/v1/databases/${encodeURIComponent(database)}/collections`, { name }).then((res) => res.data)
}

export function deleteCollection(database: string, collection: string) {
  return http.delete(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}`).then((res) => res.data)
}

export function fetchDocuments(
  database: string,
  collection: string,
  params: { page: number; pageSize: number; filter?: string; conditions?: string; logic?: string },
) {
  return http
    .get<PaginationResponse<Record<string, unknown>>>(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/documents`, {
      params: {
        page: params.page,
        page_size: params.pageSize,
        filter: params.filter,
        conditions: params.conditions,
        logic: params.logic,
      },
    })
    .then((res) => res.data)
}

export function fetchDocument(database: string, collection: string, id: string) {
  return http
    .get<DocumentEnvelope>(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/documents/${encodeURIComponent(id)}`)
    .then((res) => res.data)
}

export function createDocument(database: string, collection: string, document: Record<string, unknown>) {
  return http
    .post<DocumentEnvelope>(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/documents`, { document })
    .then((res) => res.data)
}

export function updateDocument(database: string, collection: string, id: string, document: Record<string, unknown>) {
  return http
    .put<DocumentEnvelope>(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/documents/${encodeURIComponent(id)}`, { document })
    .then((res) => res.data)
}

export function deleteDocument(database: string, collection: string, id: string) {
  return http
    .delete(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/documents/${encodeURIComponent(id)}`)
    .then((res) => res.data)
}

export function bulkDeleteDocuments(database: string, collection: string, ids: string[]) {
  return http
    .post(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/documents/bulk-delete`, { ids })
    .then((res) => res.data)
}

export function exportDocuments(database: string, collection: string, params: Record<string, unknown>) {
  const connection = getMongoSessionConnection()
  return axios.get(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/export`, {
    baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
    headers: buildMongoSessionHeaders(connection),
    params,
    responseType: 'blob',
  })
}

export function importDocuments(database: string, collection: string, file: File, format: string) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('format', format)
  return http.post(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/import`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function backupCollection(database: string, collection: string) {
  const connection = getMongoSessionConnection()
  return axios.get(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/backup`, {
    baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
    headers: buildMongoSessionHeaders(connection),
    responseType: 'blob',
  })
}

export function restoreCollection(database: string, collection: string, file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return http.post(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/restore`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function fetchIndexes(database: string, collection: string) {
  return http
    .get<ListResponse<IndexItem>>(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/indexes`)
    .then((res) => res.data)
}

export function createIndex(
  database: string,
  collection: string,
  payload: { name: string; field: string; order: number; unique: boolean; sparse: boolean },
) {
  return http
    .post(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/indexes`, payload)
    .then((res) => res.data)
}

export function deleteIndex(database: string, collection: string, name: string) {
  return http
    .delete(`/api/v1/databases/${encodeURIComponent(database)}/collections/${encodeURIComponent(collection)}/indexes/${encodeURIComponent(name)}`)
    .then((res) => res.data)
}
