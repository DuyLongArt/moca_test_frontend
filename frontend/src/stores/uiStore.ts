import { create } from 'zustand'

export type ResultsFilter = 'all' | 'pending' | 'final'

type UiState = {
  patientResultsFilter: ResultsFilter
  setPatientResultsFilter: (filter: ResultsFilter) => void
}

export const useUiStore = create<UiState>((set) => ({
  patientResultsFilter: 'all',
  setPatientResultsFilter: (patientResultsFilter) => set({ patientResultsFilter }),
}))
