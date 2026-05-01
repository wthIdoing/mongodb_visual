export interface MongoSessionConnection {
  host: string
  port: number
  database: string
  username: string
  password: string
  authSource: string
}

const STORAGE_KEY = 'mongodb_visual_connection'

export const defaultMongoConnection = (): MongoSessionConnection => ({
  host: '127.0.0.1',
  port: 27017,
  database: '',
  username: '',
  password: '',
  authSource: 'admin',
})

export function getMongoSessionConnection(): MongoSessionConnection | null {
  const raw = sessionStorage.getItem(STORAGE_KEY)
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<MongoSessionConnection>
    if (!parsed.host || !parsed.port) {
      return null
    }

    return {
      ...defaultMongoConnection(),
      ...parsed,
      port: Number(parsed.port) || 27017,
    }
  } catch {
    return null
  }
}

export function setMongoSessionConnection(connection: MongoSessionConnection) {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(connection))
}

export function clearMongoSessionConnection() {
  sessionStorage.removeItem(STORAGE_KEY)
}

export function hasMongoSessionConnection() {
  return Boolean(getMongoSessionConnection())
}

export function buildMongoSessionHeaders(connection: MongoSessionConnection | null) {
  if (!connection) {
    return {}
  }

  return {
    'X-Mongo-Host': connection.host,
    'X-Mongo-Port': String(connection.port),
    'X-Mongo-Database': connection.database,
    'X-Mongo-Username': connection.username,
    'X-Mongo-Password': connection.password,
    'X-Mongo-AuthSource': connection.authSource,
  }
}
