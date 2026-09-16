import {
  type ColumnDef,
  createPaginatedRowModel,
  createSortedRowModel,
  type RowData,
  rowPaginationFeature,
  rowSortingFeature,
  tableFeatures,
} from "@tanstack/react-table"

// react-table v9 no longer bundles every feature by default: each table opts
// into the features and row models it needs via tableFeatures(). Our tables
// only sort and paginate on the client, so we register just those two.
export const tableFeaturesConfig = tableFeatures({
  rowSortingFeature,
  rowPaginationFeature,
  sortedRowModel: createSortedRowModel(),
  paginatedRowModel: createPaginatedRowModel(),
})

export type AppTableFeatures = typeof tableFeaturesConfig

// In v9 a ColumnDef is parameterised by the table's feature set. This alias
// lets the columns.tsx files stay terse (AppColumnDef<Row>) instead of
// threading the features generic through every definition.
export type AppColumnDef<TData extends RowData, TValue = unknown> = ColumnDef<
  AppTableFeatures,
  TData,
  TValue
>
