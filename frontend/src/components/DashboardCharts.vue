<template>
  <div class="dashboard-charts">
    <div class="chart-card">
      <h4>Actividad Reciente</h4>
      <div class="chart-wrapper">
        <Bar :data="barData" :options="barOptions" />
      </div>
    </div>
    <div class="chart-card">
      <h4>Estado de Aprobación</h4>
      <div class="chart-wrapper">
        <Doughnut :data="doughnutData" :options="doughnutOptions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement
} from 'chart.js'
import { Bar, Doughnut } from 'vue-chartjs'
import { computed } from 'vue'
import { useTheme } from '@/composables/useTheme'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, ArcElement)

const { isDark } = useTheme()

const textColor = computed(() => isDark.value ? '#aeaeb2' : '#86868b')
const gridColor = computed(() => isDark.value ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)')
const surfaceColor = computed(() => isDark.value ? '#1d1d1f' : '#ffffff')

const barOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: isDark.value ? 'rgba(255,255,255,0.9)' : 'rgba(0,0,0,0.8)',
      titleColor: isDark.value ? '#000000' : '#ffffff',
      bodyColor: isDark.value ? '#000000' : '#ffffff',
      padding: 12,
      cornerRadius: 8,
      displayColors: true,
    }
  },
  scales: {
    x: {
      ticks: { color: textColor.value, font: { family: 'Inter', size: 12 } },
      grid: { display: false },
      border: { display: false }
    },
    y: {
      ticks: { color: textColor.value, font: { family: 'Inter', size: 12 } },
      grid: { color: gridColor.value, borderDash: [5, 5] },
      border: { display: false }
    }
  }
}))

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '75%',
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: textColor.value,
        padding: 20,
        font: { family: 'Inter', size: 13, weight: 'bold' as const },
        usePointStyle: true,
        pointStyle: 'circle'
      }
    },
    tooltip: {
      backgroundColor: isDark.value ? 'rgba(255,255,255,0.9)' : 'rgba(0,0,0,0.8)',
      titleColor: isDark.value ? '#000000' : '#ffffff',
      bodyColor: isDark.value ? '#000000' : '#ffffff',
      padding: 12,
      cornerRadius: 8
    }
  }
}))

const barData = computed(() => ({
  labels: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun'],
  datasets: [
    {
      label: 'Nuevos Accesos',
      backgroundColor: (context: any) => {
        const ctx = context.chart.ctx;
        const gradient = ctx.createLinearGradient(0, 0, 0, 300);
        gradient.addColorStop(0, '#f97316');
        gradient.addColorStop(1, isDark.value ? 'rgba(249,115,22,0.15)' : 'rgba(249,115,22,0.05)');
        return gradient;
      },
      hoverBackgroundColor: '#ea580c',
      borderColor: '#f97316',
      borderWidth: 2,
      borderRadius: 6,
      barThickness: 24,
      data: [40, 55, 45, 70, 90, 85]
    },
    {
      label: 'Certificados Emitidos',
      backgroundColor: (context: any) => {
        const ctx = context.chart.ctx;
        const gradient = ctx.createLinearGradient(0, 0, 0, 300);
        gradient.addColorStop(0, '#3b82f6');
        gradient.addColorStop(1, isDark.value ? 'rgba(59,130,246,0.15)' : 'rgba(59,130,246,0.05)');
        return gradient;
      },
      hoverBackgroundColor: '#2563eb',
      borderColor: '#3b82f6',
      borderWidth: 2,
      borderRadius: 6,
      barThickness: 24,
      data: [20, 30, 25, 45, 60, 50]
    }
  ]
}))

const doughnutData = computed(() => ({
  labels: ['Aprobados', 'Reprobados', 'En Progreso'],
  datasets: [
    {
      backgroundColor: ['#34c759', '#ff3b30', '#f59e0b'],
      hoverBackgroundColor: ['#28a745', '#dc3545', '#e0a800'],
      borderWidth: 4,
      borderColor: surfaceColor.value,
      hoverBorderColor: surfaceColor.value,
      data: [65, 15, 20]
    }
  ]
}))
</script>

<style scoped>
.dashboard-charts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}
.chart-card {
  background: var(--surface);
  border-radius: var(--r-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  height: 340px;
  display: flex;
  flex-direction: column;
}
.chart-card h4 {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--dark);
  margin-bottom: 16px;
}
.chart-wrapper {
  flex: 1;
  position: relative;
  min-height: 0;
}
</style>
