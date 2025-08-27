import { watch } from 'vue'
import { useColorScheme } from './useColorScheme'

/**
 * 同步 TDesign 主题与应用主题
 */
export function useTDesignTheme() {
  const { isDark } = useColorScheme()

  // 监听主题变化，同步到 TDesign
  watch(
    isDark,
    (dark) => {
      const htmlElement = document.documentElement
      
      if (dark) {
        // 深色模式
        htmlElement.setAttribute('theme-mode', 'dark')
        htmlElement.classList.add('t-dark')
      } else {
        // 浅色模式
        htmlElement.setAttribute('theme-mode', 'light')
        htmlElement.classList.remove('t-dark')
      }
    },
    { immediate: true }
  )

  return {
    isDark
  }
}