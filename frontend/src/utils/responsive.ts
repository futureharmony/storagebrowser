import { computed, ref, onMounted, onUnmounted } from "vue";

const MOBILE_BREAKPOINT = 736;
const SMALL_MOBILE_BREAKPOINT = 480;
const VERY_SMALL_MOBILE_BREAKPOINT = 400;

const windowWidth = ref(window.innerWidth);

const updateWindowWidth = () => {
  windowWidth.value = window.innerWidth;
};

export const isMobile = computed(() => windowWidth.value <= MOBILE_BREAKPOINT);
export const isSmallMobile = computed(
  () => windowWidth.value <= SMALL_MOBILE_BREAKPOINT
);
export const isVerySmallMobile = computed(
  () => windowWidth.value <= VERY_SMALL_MOBILE_BREAKPOINT
);

export const responsiveClass = (desktopClass: string, mobileClass: string) => {
  return isMobile.value ? mobileClass : desktopClass;
};

export const useResponsive = () => {
  onMounted(() => {
    window.addEventListener("resize", updateWindowWidth);
  });

  onUnmounted(() => {
    window.removeEventListener("resize", updateWindowWidth);
  });

  return {
    isMobile,
    isSmallMobile,
    isVerySmallMobile,
    responsiveClass,
  };
};
