<template>
  <div class="min-h-screen bg-[#0A0A0B] flex flex-col items-center justify-center p-4 relative overflow-hidden">
    <!-- Subtle Background Glow -->
    <div class="absolute w-[500px] h-[500px] bg-[#7B96F5]/5 rounded-full blur-3xl pointer-events-none -top-40 -left-40"></div>
    <div class="absolute w-[400px] h-[400px] bg-[#3ECF8E]/5 rounded-full blur-3xl pointer-events-none -bottom-20 -right-20"></div>

    <!-- Login Card -->
    <div class="w-full max-w-md bg-[#151517] border border-[#26262A] rounded-2xl p-8 shadow-2xl relative z-10">
      <!-- Top Icon & Header -->
      <div class="flex flex-col items-center text-center">
        <div class="w-14 h-14 rounded-2xl bg-[#0082FF]/10 border border-[#0082FF]/30 flex items-center justify-center p-2.5 shadow-lg shadow-[#0082FF]/20 mb-4">
          <img src="../assets/logo-sanoc-mark.svg" alt="SANOC Logo" class="w-full h-full object-contain" />
        </div>
        <h1 class="text-2xl font-black text-white tracking-wider">SANOC</h1>
        <p class="text-xs text-gray-400 mt-1 font-medium">Enterprise Infrastructure &amp; Network Monitoring</p>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="mt-8 space-y-5">
        <!-- Username / Email Field -->
        <div class="space-y-1.5">
          <label class="block text-[10px] font-mono uppercase tracking-wider text-gray-400 font-semibold">Username / Email</label>
          <div class="relative">
            <User class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              v-model="username"
              type="text"
              required
              placeholder="Enter your username"
              class="w-full bg-[#18181B] border rounded-lg pl-10 pr-4 py-2.5 text-xs text-gray-200 placeholder-gray-500 focus:outline-none transition-colors"
              :class="errorMessage ? 'border-red-500/80 focus:border-red-500' : 'border-[#26262A] focus:border-[#7B96F5]'"
            />
          </div>
        </div>

        <!-- Password Field -->
        <div class="space-y-1.5">
          <label class="block text-[10px] font-mono uppercase tracking-wider text-gray-400 font-semibold">Password</label>
          <div class="relative">
            <Lock class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              required
              placeholder="Enter your password"
              class="w-full bg-[#18181B] border rounded-lg pl-10 pr-10 py-2.5 text-xs text-gray-200 placeholder-gray-500 focus:outline-none transition-colors"
              :class="errorMessage ? 'border-red-500/80 focus:border-red-500' : 'border-[#26262A] focus:border-[#7B96F5]'"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <Eye v-if="!showPassword" class="w-4 h-4" />
              <EyeOff v-else class="w-4 h-4" />
            </button>
          </div>

          <!-- Inline Red Error State -->
          <div v-if="errorMessage" class="flex items-center gap-1.5 text-xs text-red-400 pt-1 font-mono">
            <AlertCircle class="w-3.5 h-3.5 shrink-0" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <!-- Remember Me Checkbox -->
        <div class="flex items-center justify-between text-xs">
          <label class="flex items-center gap-2 text-gray-400 cursor-pointer hover:text-gray-300">
            <input
              v-model="rememberMe"
              type="checkbox"
              class="rounded border-[#26262A] bg-[#18181B] text-[#7B96F5] focus:ring-0 accent-[#7B96F5]"
            />
            <span>Remember me</span>
          </label>
        </div>

        <!-- Google reCAPTCHA v2 Checkbox Container -->
        <div v-show="recaptchaSiteKey" class="flex justify-center my-3 min-h-[78px]">
          <div ref="recaptchaContainer"></div>
        </div>

        <!-- Sign In Button -->
        <button
          type="submit"
          :disabled="isSubmitting"
          class="w-full py-2.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-lg shadow-[#7B96F5]/25 transition-all flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
        >
          <span v-if="!isSubmitting">Sign In</span>
          <span v-else class="animate-pulse">Authenticating...</span>
        </button>
      </form>
    </div>

    <!-- MFA 6-Digit Verification Modal -->
    <div v-if="showMFAModal" class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50 animate-fadeIn">
      <div class="bg-[#151517] border border-[#7B96F5]/30 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-5">
        <div class="flex items-center gap-3 border-b border-[#26262A] pb-4">
          <div class="p-2.5 bg-[#7B96F5]/10 text-[#7B96F5] rounded-xl border border-[#7B96F5]/30">
            <ShieldCheck class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-sm font-bold text-white font-mono">Two-Factor Authentication (MFA)</h3>
            <p class="text-xs text-gray-400 font-mono mt-0.5">Enter the 6-digit passcode from your Authenticator app</p>
          </div>
        </div>

        <form @submit.prevent="handleVerifyMFA" class="space-y-5">
          <div class="space-y-2">
            <label class="block text-center text-[10px] font-mono uppercase tracking-wider text-gray-400 font-semibold">
              6-Digit Authenticator Passcode
            </label>
            <OtpInput
              v-model="mfaCode"
              :error="!!mfaError"
              :disabled="isVerifyingMFA"
              @complete="handleVerifyMFA"
            />
          </div>

          <div v-if="mfaError" class="flex items-center justify-center gap-1.5 text-xs text-red-400 font-mono">
            <AlertCircle class="w-3.5 h-3.5 shrink-0" />
            <span>{{ mfaError }}</span>
          </div>

          <div class="flex items-center gap-3 pt-2">
            <button
              type="button"
              @click="showMFAModal = false"
              class="flex-1 py-2.5 bg-[#26262A] hover:bg-[#333338] text-gray-300 rounded-lg text-xs font-mono font-bold transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isVerifyingMFA || mfaCode.length !== 6"
              class="flex-1 py-2.5 bg-[#7B96F5] hover:bg-[#95ABF7] disabled:opacity-50 text-white rounded-lg text-xs font-mono font-bold transition-colors flex items-center justify-center gap-2"
            >
              <span v-if="!isVerifyingMFA">Verify Passcode</span>
              <span v-else class="animate-pulse">Verifying...</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Footer Tag -->
    <div class="mt-8 text-center text-[10px] font-mono text-gray-500 tracking-wider space-y-1">
      <div class="text-gray-400">SANOC Infrastructure Monitoring</div>
      <div class="text-gray-500">© SANOC Team — UTB 2026.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '../stores/authStore';
import OtpInput from '../components/common/OtpInput.vue';
import { User, Lock, Eye, EyeOff, AlertCircle, ShieldCheck } from 'lucide-vue-next';

const authStore = useAuthStore();

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

onMounted(() => {
  if (!recaptchaSiteKey) return;

  const renderWidget = () => {
    if (!recaptchaContainer.value || recaptchaWidgetId !== null) return;
    const grecaptcha = (window as any).grecaptcha;
    if (grecaptcha && typeof grecaptcha.render === 'function') {
      try {
        recaptchaWidgetId = grecaptcha.render(recaptchaContainer.value, {
          sitekey: recaptchaSiteKey,
          theme: 'dark',
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
