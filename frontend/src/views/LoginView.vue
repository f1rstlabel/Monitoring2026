<template>
  <div class="min-h-dvh bg-main flex flex-col items-center justify-center p-4 sm:p-6 relative overflow-hidden transition-colors duration-200">
    <!-- Subtle Background Glow -->
    <div class="absolute w-[500px] h-[500px] bg-brand-periwinkle/5 rounded-full blur-3xl pointer-events-none -top-40 -left-40"></div>
    <div class="absolute w-[400px] h-[400px] bg-brand-periwinkle/5 rounded-full blur-3xl pointer-events-none -bottom-20 -right-20"></div>

    <!-- Theme Control -->
    <button
      type="button"
      @click="themeStore.toggleTheme"
      class="absolute right-4 top-4 sm:right-6 sm:top-6 z-20 p-2 rounded-lg bg-surface border border-subtle text-text-secondary hover:text-text-main hover:bg-card transition-colors cursor-pointer"
      :aria-label="themeStore.currentTheme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'"
      title="Toggle Theme"
    >
      <Sun v-if="themeStore.currentTheme === 'dark'" class="w-4 h-4" />
      <Moon v-else class="w-4 h-4" />
    </button>

    <!-- Login Card -->
    <main class="w-full max-w-md bg-surface border border-subtle rounded-2xl p-7 sm:p-8 shadow-2xl shadow-brand-periwinkle/5 relative z-10">
      <!-- Top Icon & Header -->
      <div class="flex flex-col items-center text-center">
        <div class="w-16 h-16 rounded-2xl bg-card border border-subtle flex items-center justify-center shadow-lg shadow-brand-periwinkle/10 mb-4 overflow-hidden relative">
          <img
            v-if="settingStore.branding.logoUrl"
            :src="settingStore.branding.logoUrl"
            alt="Logo"
            class="w-full h-full block"
            :class="(settingStore.branding.logoFit || 'cover') === 'cover' ? 'object-cover' : 'object-contain p-2'"
          />
          <img
            v-else
            src="../assets/logo-sanoc-mark.svg"
            alt="SANOC Logo"
            class="w-full h-full object-contain p-2 block"
          />
        </div>
        <h1 class="text-2xl sm:text-3xl font-black text-text-main tracking-[0.16em]">{{ settingStore.branding.appTitle || 'SANOC' }}</h1>
        <p class="text-xs text-text-secondary mt-1 font-medium max-w-xs leading-relaxed">{{ settingStore.branding.appSubtitle || 'Enterprise Infrastructure & Network Monitoring' }}</p>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="mt-8 space-y-5">
        <!-- Username / Email Field -->
        <div class="space-y-1.5">
          <label class="block text-[10px] font-mono uppercase tracking-wider text-text-secondary font-semibold">Username / Email</label>
          <div class="relative">
            <User class="w-4 h-4 text-text-secondary absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              v-model="username"
              type="text"
              required
              placeholder="Enter your username"
              class="w-full bg-card border rounded-lg pl-10 pr-4 py-2.5 text-xs text-text-main placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-brand-periwinkle/20 transition-colors"
              :class="errorMessage ? 'border-status-down/80 focus:border-status-down' : 'border-subtle focus:border-brand-periwinkle'"
            />
          </div>
        </div>

        <!-- Password Field -->
        <div class="space-y-1.5">
          <label class="block text-[10px] font-mono uppercase tracking-wider text-text-secondary font-semibold">Password</label>
          <div class="relative">
            <Lock class="w-4 h-4 text-text-secondary absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              required
              placeholder="Enter your password"
              class="w-full bg-card border rounded-lg pl-10 pr-10 py-2.5 text-xs text-text-main placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-brand-periwinkle/20 transition-colors"
              :class="errorMessage ? 'border-status-down/80 focus:border-status-down' : 'border-subtle focus:border-brand-periwinkle'"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary hover:text-text-main transition-colors cursor-pointer"
              :aria-label="showPassword ? 'Hide password' : 'Show password'"
            >
              <Eye v-if="!showPassword" class="w-4 h-4" />
              <EyeOff v-else class="w-4 h-4" />
            </button>
          </div>

          <!-- Inline Red Error State -->
          <div v-if="errorMessage" class="flex items-center gap-1.5 text-xs text-status-down pt-1 font-mono" role="alert">
            <AlertCircle class="w-3.5 h-3.5 shrink-0" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <!-- Remember Me Checkbox -->
        <div class="flex items-center justify-between text-xs">
          <label class="flex items-center gap-2 text-text-secondary cursor-pointer hover:text-text-secondary">
            <input
              v-model="rememberMe"
              type="checkbox"
              class="rounded border-subtle bg-card text-brand-periwinkle focus:ring-2 focus:ring-brand-periwinkle/20"
            />
            <span>Remember me</span>
          </label>
        </div>

        <!-- Google reCAPTCHA v2 Checkbox Container -->
        <div v-if="recaptchaSiteKey" class="flex justify-center my-3 min-h-[78px] overflow-hidden rounded-lg">
          <div ref="recaptchaContainer" class="max-w-full"></div>
        </div>

        <!-- Sign In Button -->
        <button
          type="submit"
          :disabled="isSubmitting"
          class="w-full py-2.5 rounded-lg bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white font-semibold text-xs shadow-lg shadow-brand-periwinkle/25 transition-all flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
        >
          <span v-if="!isSubmitting">Sign In</span>
          <span v-else class="animate-pulse">Authenticating...</span>
        </button>
      </form>
    </main>

    <!-- MFA 6-Digit Verification Modal -->
    <div v-if="showMFAModal" class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50 animate-fadeIn">
      <div class="bg-surface border border-brand-periwinkle/30 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-5">
        <div class="flex items-center gap-3 border-b border-subtle pb-4">
          <div class="p-2.5 bg-brand-periwinkle/10 text-brand-periwinkle rounded-xl border border-brand-periwinkle/30">
            <ShieldCheck class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-sm font-bold text-text-main font-mono">Two-Factor Authentication (MFA)</h3>
            <p class="text-xs text-text-secondary font-mono mt-0.5">Enter the 6-digit passcode from your Authenticator app</p>
          </div>
        </div>

        <form @submit.prevent="handleVerifyMFA" class="space-y-5">
          <div class="space-y-2">
            <label class="block text-center text-[10px] font-mono uppercase tracking-wider text-text-secondary font-semibold">
              6-Digit Authenticator Passcode
            </label>
            <OtpInput
              v-model="mfaCode"
              :error="!!mfaError"
              :disabled="isVerifyingMFA"
              @complete="handleVerifyMFA"
            />
          </div>

          <div v-if="mfaError" class="flex items-center justify-center gap-1.5 text-xs text-status-down font-mono" role="alert">
            <AlertCircle class="w-3.5 h-3.5 shrink-0" />
            <span>{{ mfaError }}</span>
          </div>

          <div class="flex items-center gap-3 pt-2">
            <button
              type="button"
              @click="showMFAModal = false"
              class="flex-1 py-2.5 bg-subtle hover:bg-hover text-text-secondary rounded-lg text-xs font-mono font-bold transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isVerifyingMFA || mfaCode.length !== 6"
              class="flex-1 py-2.5 bg-brand-periwinkle hover:bg-brand-periwinkle-hover disabled:opacity-50 text-white rounded-lg text-xs font-mono font-bold transition-colors flex items-center justify-center gap-2"
            >
              <span v-if="!isVerifyingMFA">Verify Passcode</span>
              <span v-else class="animate-pulse">Verifying...</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Footer Tag -->
    <div class="mt-8 text-center text-[10px] font-mono text-text-muted tracking-wider space-y-1">
      <div class="text-text-secondary">SANOC Infrastructure Monitoring</div>
      <div class="text-text-muted">© SANOC Team — UTB 2026.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue';
import { useAuthStore } from '../stores/authStore';
import { useSettingStore } from '../stores/settingStore';
import { useThemeStore } from '../stores/themeStore';
import OtpInput from '../components/common/OtpInput.vue';
import { User, Lock, Eye, EyeOff, AlertCircle, ShieldCheck, Sun, Moon } from 'lucide-vue-next';

const authStore = useAuthStore();
const settingStore = useSettingStore();
const themeStore = useThemeStore();

const username = ref('');
const password = ref('');
const showPassword = ref(false);
const rememberMe = ref(true);
const errorMessage = ref('');
const isSubmitting = ref(false);

const showMFAModal = ref(false);
const mfaToken = ref('');
const mfaCode = ref('');
const mfaError = ref('');
const isVerifyingMFA = ref(false);

const recaptchaSiteKey = ((import.meta as any).env?.VITE_RECAPTCHA_SITE_KEY as string) || '';
const recaptchaContainer = ref<HTMLElement | null>(null);
const recaptchaToken = ref('');
let recaptchaWidgetId: number | null = null;
let renderRecaptcha: (() => void) | null = null;

onMounted(() => {
  settingStore.fetchBranding();
  if (!recaptchaSiteKey) return;

  const renderWidget = () => {
    if (!recaptchaContainer.value || recaptchaWidgetId !== null) return;
    const grecaptcha = (window as any).grecaptcha;
    if (grecaptcha && typeof grecaptcha.render === 'function') {
      try {
        recaptchaWidgetId = grecaptcha.render(recaptchaContainer.value, {
          sitekey: recaptchaSiteKey,
          theme: themeStore.currentTheme,
          callback: (token: string) => {
            recaptchaToken.value = token;
            errorMessage.value = '';
          },
          'expired-callback': () => {
            recaptchaToken.value = '';
          },
          'error-callback': () => {
            recaptchaToken.value = '';
          }
        });
      } catch (err) {
        console.warn('[reCAPTCHA v2] Render notice:', err);
      }
    }
  };

  renderRecaptcha = renderWidget;

  if ((window as any).grecaptcha && (window as any).grecaptcha.render) {
    renderWidget();
  } else {
    (window as any).onloadRecaptchaCallback = () => {
      renderWidget();
    };
    if (!document.getElementById('recaptcha-script')) {
      const script = document.createElement('script');
      script.id = 'recaptcha-script';
      script.src = 'https://www.google.com/recaptcha/api.js?onload=onloadRecaptchaCallback&render=explicit';
      script.async = true;
      script.defer = true;
      document.head.appendChild(script);
    }
  }
});

watch(() => themeStore.currentTheme, async () => {
  if (!recaptchaSiteKey || !recaptchaContainer.value || recaptchaWidgetId === null) return;

  const grecaptcha = (window as any).grecaptcha;
  if (!grecaptcha) return;

  try {
    grecaptcha.reset(recaptchaWidgetId);
  } catch (e) {}

  recaptchaWidgetId = null;
  recaptchaToken.value = '';
  recaptchaContainer.value.innerHTML = '';
  await nextTick();
  renderRecaptcha?.();
});

function resetRecaptcha() {
  const grecaptcha = (window as any).grecaptcha;
  if (grecaptcha && recaptchaWidgetId !== null) {
    try {
      grecaptcha.reset(recaptchaWidgetId);
    } catch (e) {}
  }
  recaptchaToken.value = '';
}

async function handleLogin() {
  errorMessage.value = '';

  if (recaptchaSiteKey && !recaptchaToken.value) {
    errorMessage.value = 'Harap centang verifikasi reCAPTCHA ("Saya bukan robot") terlebih dahulu.';
    return;
  }

  isSubmitting.value = true;

  try {
    const res = await authStore.login(username.value, password.value, rememberMe.value, recaptchaToken.value);
    if (res.requireMFA) {
      mfaToken.value = res.mfaToken;
      showMFAModal.value = true;
    } else if (res.success) {
      window.location.href = '/dashboard';
    } else {
      errorMessage.value = res.message || 'Invalid username or password';
      resetRecaptcha();
    }
  } catch (e: any) {
    errorMessage.value = e.response?.data?.message || 'Invalid username or password';
    resetRecaptcha();
  } finally {
    isSubmitting.value = false;
  }
}

async function handleVerifyMFA() {
  const cleanCode = mfaCode.value.replace(/\s+|-/g, '');
  if (!cleanCode || cleanCode.length !== 6) {
    mfaError.value = 'Kode verifikasi MFA harus terdiri dari 6 digit angka';
    return;
  }

  mfaError.value = '';
  isVerifyingMFA.value = true;

  try {
    const res = await authStore.verifyLoginMFA(mfaToken.value, cleanCode);
    if (res.success) {
      showMFAModal.value = false;
      mfaCode.value = '';
      window.location.href = '/dashboard';
    } else {
      mfaError.value = res.message || 'Kode MFA tidak valid atau telah kadaluwarsa';
    }
  } catch (e: any) {
    mfaError.value = e.response?.data?.message || 'Gagal memverifikasi kode MFA';
  } finally {
    isVerifyingMFA.value = false;
  }
}
</script>
