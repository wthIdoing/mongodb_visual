import { defineStore } from 'pinia'

import { fetchCollections, fetchConnectionInfo, fetchDatabases } from '../api/mongo'
import type { CollectionItem, ConnectionInfo } from '../api/types'

export const useMongoStore = defineStore('mongo', {
  state: () => ({
    connection: null as ConnectionInfo | null,
    databases: [] as string[],
    collectionsByDatabase: {} as Record<string, CollectionItem[]>,
    loadingTree: false,
  }),
  actions: {
    async loadConnection() {
      this.connection = await fetchConnectionInfo()
    },
    async loadDatabases() {
      this.loadingTree = true
      try {
        const data = await fetchDatabases()
        this.databases = data.items
      } finally {
        this.loadingTree = false
      }
    },
    async loadCollections(database: string) {
      const data = await fetchCollections(database)
      this.collectionsByDatabase[database] = data.items
      return data.items
    },
  },
})
