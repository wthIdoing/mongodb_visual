<template>
  <div class="sidebar-tree">
    <div class="sidebar-tree__header">
      <div>
        <p class="eyebrow">{{ copy.sidebar.eyebrow }}</p>
        <h2>{{ copy.sidebar.title }}</h2>
      </div>
      <el-button text class="sidebar-tree__refresh" @click="$emit('refresh')">{{ copy.common.refresh }}</el-button>
    </div>

    <el-scrollbar>
      <div v-loading="loading" class="sidebar-tree__body">
        <div v-for="database in databases" :key="database" class="sidebar-tree__database">
          <div class="sidebar-tree__row">
            <button
              class="sidebar-tree__database-button"
              :class="{ 'sidebar-tree__database-button--active': selectedDatabase === database }"
              @click="$emit('toggle-database', database)"
            >
              <span class="sidebar-tree__database-name">{{ database }}</span>
              <span class="sidebar-tree__database-meta">{{ copy.sidebar.collectionCount(collections[database]?.length ?? 0) }}</span>
            </button>
            <el-button text size="small" class="sidebar-tree__create" @click="$emit('create-collection', database)">{{ copy.sidebar.createCollection }}</el-button>
          </div>

          <div v-if="collections[database]?.length" class="sidebar-tree__collections">
            <button
              v-for="collection in collections[database]"
              :key="`${database}-${collection}`"
              class="sidebar-tree__collection-button"
              :class="{
                'sidebar-tree__collection-button--active':
                  selectedDatabase === database && selectedCollection === collection,
              }"
              @click="$emit('select-collection', { database, collection })"
            >
              <span>{{ collection }}</span>
              <span class="sidebar-tree__collection-tag">{{ copy.sidebar.collectionTag }}</span>
            </button>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { copy } from '../copy'

defineProps<{
  loading: boolean
  databases: string[]
  collections: Record<string, string[]>
  selectedDatabase: string
  selectedCollection: string
}>()

defineEmits<{
  refresh: []
  'toggle-database': [database: string]
  'select-collection': [payload: { database: string; collection: string }]
  'create-collection': [database: string]
}>()
</script>
