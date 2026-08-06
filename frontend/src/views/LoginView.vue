<template>
  <div class="min-h-screen bg-[#0A0A0B] flex flex-col items-center justify-center p-4 relative overflow-hidden">
    <!-- Subtle Background Glow -->
    <div class="absolute w-[500px] h-[500px] bg-[#7B96F5]/5 rounded-full blur-3xl pointer-events-none -top-40 -left-40"></div>
    <div class="absolute w-[400px] h-[400px] bg-[#3ECF8E]/5 rounded-full blur-3xl pointer-events-none -bottom-20 -right-20"></div>

    <!-- Login Card -->
    <div class="w-full max-w-md bg-[#151517] border border-[#26262A] rounded-2xl p-8 shadow-2xl relative z-10">
      <!-- Top Icon & Header -->
      <div class="flex flex-col items-center text-center">
        <div class="w-14 h-14 rounded-2xl bg-[#7B96F5]/15 border border-[#7B96F5]/30 flex items-center justify-center text-[#7B96F5] shadow-lg shadow-[#7B96F5]/10 mb-4">
          <Shield class="w-7 h-7" />
        </div>
        <h1 class="text-xl font-bold text-white tracking-tight">Sistem Monitoring Server & Jaringan</h1>
        <p class="text-xs text-gray-400 mt-1 font-medium">Biro Umum, Sekretariat Daerah Provinsi Jawa Barat</p>
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

        <!-- Sign In Button -->
        <button
          type="submit"
          :disabled="isSubmitting"
          class="w-full py-2.5 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs shadow-lg shadow-[#7B96F5]/25 transition-all flex items-center justify-center gap-2"
        >
          <span v-if="!isSubmitting">Sign In</span>
          <span v-else class="animate-pulse">Authenticating...</span>
        </button>
      </form>

      <!-- Forgot Password Link -->
      <div class="mt-6 text-center">
        <a href="#" @click.prevent="isForgotModalOpen = true" class="text-xs text-gray-500 hover:text-[#7B96F5] transition-colors">
          Forgot password?
        </a>
      </div>
    </div>

    <!-- Footer Tag -->
    <div class="mt-8 text-center text-[10px] font-mono text-gray-600 tracking-wider">
      v2.4.1-stable // NOC SECURE LOGIN
    </div>

    <!-- Forgot Password Modal -->
    <Modal :is-open="isForgotModalOpen" title="Reset Password" @close="isForgotModalOpen = false">
      <template #default>
        <form @submit.prevent="handleForgotPassword" class="space-y-4 text-xs">
          <div class="space-y-1.5">
            <label class="block font-mono uppercase text-[10px] text-gray-400">Email Address *</label>
            <input
              v-model="forgotEmail"
              type="email"
              required
              placeholder="e.g. admin.noc@jabarprov.go.id"
              class="w-full bg-[#18181B] border border-[#26262A] rounded-lg px-3 py-2 text-gray-200 focus:outline-none focus:border-[#7B96F5]"
            />
          </div>

          <div v-if="forgotMessage" class="p-2.5 rounded bg-[#34D399]/10 border border-[#34D399]/30 text-[#34D399] text-xs font-mono">
            {{ forgotMessage }}
          </div>
        </form>
      </template>

      <template #footer>
        <button
          @click="isForgotModalOpen = false"
          class="px-4 py-2 rounded-lg border border-[#26262A] text-gray-400 hover:text-gray-200 text-xs"
        >
          Close
        </button>
        <button
          @click="handleForgotPassword"
          :disabled="!forgotEmail || isSendingForgot"
          class="px-5 py-2 rounded-lg bg-[#7B96F5] hover:bg-[#95ABF7] text-white font-semibold text-xs flex items-center gap-2 disabled:opacity-50"
        >
          <span v-if="!isSendingForgot">Send Reset Link</span>
          <span v-else class="animate-pulse">Sending...</span>
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/authStore';
import { authApi } from '../api';
import Modal from '../components/common/Modal.vue';
import { Shield, User, Lock, Eye, EyeOff, AlertCircle } from 'lucide-vue-next';

const router = useRouter();
const authStore = useAuthStore();

const username = ref('');
const password = ref('');
const showPassword = ref(false);
const rememberMe = ref(true);
const errorMessage = ref('');
const isSubmitting = ref(false);

const isForgotModalOpen = ref(false);
const forgotEmail = ref('');
const forgotMessage = ref('');
const isSendingForgot = ref(false);

async function handleLogin() {
  errorMessage.value = '';
  isSubmitting.value = true;

  try {
    const res = await authStore.login(username.value, password.value, rememberMe.value);
    if (res.success) {
      router.push('/dashboard');
    } else {
      errorMessage.value = res.message || 'Invalid username or password';
    }
  } catch (e: any) {
    errorMessage.value = e.response?.data?.message || 'Invalid username or password';
  } finally {
    isSubmitting.value = false;
  }
}

async function handleForgotPassword() {
  if (!forgotEmail.value) return;
  isSendingForgot.value = true;
  forgotMessage.value = '';
  try {
    const res = await authApi.forgotPassword(forgotEmail.value);
    forgotMessage.value = res.message || 'Password reset link sent to your email';
  } catch (e: any) {
    forgotMessage.value = e.response?.data?.message || 'Failed to send reset link';
  } finally {
    isSendingForgot.value = false;
  }
}
</script>
