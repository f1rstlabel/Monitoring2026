<template>
  <div class="space-y-6 max-w-4xl mx-auto pb-12">
    <!-- Header Page Banner -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-text-main flex items-center gap-2.5">
          <User class="w-6 h-6 text-brand-periwinkle" />
          User Profile
        </h1>
        <p class="text-xs text-text-secondary mt-1">
          Account details, security settings, and profile management for SANOC Monitoring
        </p>
      </div>

      <div class="flex items-center gap-2">
        <span
          class="px-2.5 py-1 rounded text-xs font-mono font-semibold uppercase border"
          :class="[
            authStore.user.role === 'admin'
              ? 'bg-purple-500/10 text-purple-400 border-purple-500/30'
              : authStore.user.role === 'pimpinan'
              ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
              : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
          ]"
        >
          Role: {{ authStore.user.role }}
        </span>
      </div>
    </div>

    <!-- Alert / Toast Banner -->
    <div
      v-if="toastMessage"
      class="p-4 rounded-xl border flex items-center justify-between text-xs font-mono transition-all animate-fadeIn"
      :class="[
        toastType === 'success'
          ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400'
          : 'bg-red-500/10 border-red-500/30 text-red-400'
      ]"
    >
      <div class="flex items-center gap-2">
        <CheckCircle v-if="toastType === 'success'" class="w-4 h-4 shrink-0" />
        <AlertTriangle v-else class="w-4 h-4 shrink-0" />
        <span>{{ toastMessage }}</span>
      </div>
      <button @click="toastMessage = ''" class="hover:opacity-80">
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- Main Profile Form Card -->
    <div class="bg-surface border border-subtle rounded-xl p-6 space-y-6 shadow-xl">
      <!-- Section 1: Profile Picture & Avatar -->
      <div class="border-b border-subtle pb-6 space-y-4">
        <h2 class="text-sm font-bold text-text-main uppercase tracking-wider font-mono flex items-center gap-2">
          <Camera class="w-4 h-4 text-brand-periwinkle" />
          Profile Avatar &amp; Picture
        </h2>

        <div class="flex flex-col sm:flex-row items-center gap-6">
          <!-- Current Avatar Preview with Change Photo Overlay -->
          <div class="relative group cursor-pointer shrink-0" @click="triggerFilePicker" title="Click to upload new profile photo">
            <img
              :src="form.avatarUrl || 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'"
              alt="Profile Avatar"
              @error="(e: Event) => (e.target as HTMLImageElement).src = 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'"
              class="w-24 h-24 rounded-full object-cover border-2 border-brand-periwinkle/50 shadow-lg ring-4 ring-surface group-hover:opacity-80 transition-all"
            />
            <div class="absolute inset-0 bg-black/60 rounded-full flex flex-col items-center justify-center text-text-main opacity-0 group-hover:opacity-100 transition-opacity text-[10px] font-mono font-bold">
              <Camera class="w-5 h-5 mb-0.5" />
              <span>Change Photo</span>
            </div>
            <!-- Camera badge icon at bottom right -->
            <div class="absolute bottom-0 right-0 p-1.5 bg-brand-periwinkle text-white rounded-full shadow-md border-2 border-surface">
              <Camera class="w-3.5 h-3.5" />
            </div>
          </div>

          <!-- Direct Change Photo Action & Helper Text -->
          <div class="flex-1 space-y-2 w-full text-center sm:text-left">
            <div>
              <h3 class="text-xs font-bold text-text-main font-mono">Profile Photo</h3>
              <p class="text-[11px] text-text-secondary font-mono mt-0.5">
                Click on the avatar image to upload a photo directly from your computer (JPG, PNG, WEBP max 10MB).
              </p>
            </div>

            <div class="flex items-center gap-3 justify-center sm:justify-start pt-1">
              <input
                ref="fileInputRef"
                type="file"
                accept="image/jpeg,image/png,image/webp,image/gif"
                class="hidden"
                @change="handleFileUpload"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Section 2: Account Details (Username & Email) -->
      <div class="border-b border-subtle pb-6 space-y-4">
        <h2 class="text-sm font-bold text-text-main uppercase tracking-wider font-mono flex items-center gap-2">
          <Mail class="w-4 h-4 text-status-up" />
          Account Information
        </h2>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">Username *</label>
            <input
              v-model="form.username"
              type="text"
              required
              placeholder="e.g. admin.noc"
              class="w-full bg-card border border-subtle focus:border-status-up rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">Full Name *</label>
            <input
              v-model="form.name"
              type="text"
              required
              placeholder="e.g. Budi Santoso"
              class="w-full bg-card border border-subtle focus:border-status-up rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">Email Address *</label>
            <input
              v-model="form.email"
              type="email"
              required
              placeholder="e.g. user@jabarprov.go.id"
              class="w-full bg-card border border-subtle focus:border-status-up rounded-lg px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
            />
          </div>
        </div>
      </div>

      <!-- Section 3: Security & Change Password (Dual Method: Old Password or Email OTP) -->
      <div class="space-y-4 border-b border-subtle pb-6">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-bold text-text-main uppercase tracking-wider font-mono flex items-center gap-2">
              <Lock class="w-4 h-4 text-amber-400" />
              Security &amp; Password Management
            </h2>
            <p class="text-[11px] text-text-secondary font-mono mt-0.5">
              Select your preferred password update method:
            </p>
          </div>

          <!-- Method Selector Switcher -->
          <div class="inline-flex bg-card border border-subtle rounded-xl p-1 shadow-sm">
            <button
              type="button"
              @click="pwdChangeMethod = 'current_password'"
              class="px-3 py-1.5 rounded-lg text-xs font-mono font-semibold transition-all flex items-center gap-1.5 cursor-pointer"
              :class="pwdChangeMethod === 'current_password' ? 'bg-brand-periwinkle text-white shadow-xs' : 'text-text-secondary hover:text-text-main'"
            >
              <KeyRound class="w-3.5 h-3.5" />
              <span>Current Password</span>
            </button>
            <button
              type="button"
              @click="pwdChangeMethod = 'email_otp'"
              class="px-3 py-1.5 rounded-lg text-xs font-mono font-semibold transition-all flex items-center gap-1.5 cursor-pointer"
              :class="pwdChangeMethod === 'email_otp' ? 'bg-brand-periwinkle text-white shadow-xs' : 'text-text-secondary hover:text-text-main'"
            >
              <Mail class="w-3.5 h-3.5" />
              <span>Email Verification (OTP)</span>
            </button>
          </div>
        </div>

        <!-- METODE A: Menggunakan Kata Sandi Saat Ini -->
        <div v-if="pwdChangeMethod === 'current_password'" class="grid grid-cols-1 sm:grid-cols-2 gap-4 animate-fadeIn">
          <div>
            <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">Current Password</label>
            <input
              v-model="form.currentPassword"
              type="password"
              placeholder="Enter current password"
              class="w-full bg-card border border-subtle focus:border-amber-400 rounded-xl px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">New Password (Min 12 Characters)</label>
            <input
              v-model="form.newPassword"
              type="password"
              placeholder="At least 12 characters"
              class="w-full bg-card border border-subtle focus:border-amber-400 rounded-xl px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
            />
            <div v-if="form.newPassword" class="mt-2 space-y-1">
              <div class="flex items-center justify-between text-[10px] font-mono">
                <span class="text-text-secondary">Strength: {{ passwordStrengthText }}</span>
                <span :class="passwordLengthValid ? 'text-emerald-400' : 'text-red-400'">{{ form.newPassword.length }}/12</span>
              </div>
              <div class="w-full h-1 bg-subtle rounded-full overflow-hidden">
                <div
                  class="h-full transition-all duration-300"
                  :class="[
                    passwordScore <= 1 ? 'w-1/4 bg-red-500' :
                    passwordScore === 2 ? 'w-2/4 bg-amber-500' :
                    passwordScore === 3 ? 'w-3/4 bg-blue-500' : 'w-full bg-emerald-500'
                  ]"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- METODE B: Menggunakan Verifikasi Email (OTP) -->
        <div v-else class="p-4 bg-card border border-subtle rounded-2xl space-y-4 animate-fadeIn">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 bg-surface rounded-xl border border-subtle">
            <div class="flex items-center gap-2.5">
              <div class="p-2 rounded-lg bg-brand-periwinkle/10 text-brand-periwinkle">
                <Mail class="w-4 h-4" />
              </div>
              <div>
                <span class="text-[10px] font-mono uppercase text-text-secondary block font-semibold">Registered Account Email</span>
                <strong class="text-xs font-mono text-text-main">{{ form.email }}</strong>
              </div>
            </div>

            <button
              type="button"
              @click="handleSendProfileResetOTP"
              :disabled="profileOTPCountdown > 0 || isSendingProfileOTP"
              class="px-3.5 py-1.5 bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white text-xs font-mono font-bold rounded-xl shadow-md shadow-brand-periwinkle/20 disabled:opacity-50 flex items-center gap-1.5 self-start sm:self-auto cursor-pointer transition-colors"
            >
              <Send class="w-3.5 h-3.5" :class="isSendingProfileOTP ? 'animate-spin' : ''" />
              <span v-if="profileOTPCountdown > 0">Resend in ({{ profileOTPCountdown }}s)</span>
              <span v-else>{{ isSendingProfileOTP ? 'Sending...' : 'Send Verification Code to Email' }}</span>
            </button>
          </div>

          <div class="p-3 bg-surface rounded-xl border border-subtle space-y-2 text-center">
            <label class="block text-xs font-semibold text-text-secondary font-mono">
              Enter 6-Digit Email Verification Code (1 Digit Per Box) *
            </label>
            <OtpInput
              v-model="otpResetCode"
              :length="6"
              :auto-focus="false"
            />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">New Password * (Min 8 Characters)</label>
              <input
                v-model="otpResetNewPassword"
                type="password"
                placeholder="At least 8 characters"
                class="w-full bg-card border border-subtle focus:border-brand-periwinkle rounded-xl px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
              />
            </div>

            <div>
              <label class="block text-xs font-medium text-text-secondary mb-1 font-mono">Confirm New Password *</label>
              <input
                v-model="otpResetConfirmPassword"
                type="password"
                placeholder="Repeat new password"
                class="w-full bg-card border border-subtle focus:border-brand-periwinkle rounded-xl px-3 py-2 text-xs text-text-main focus:outline-none font-mono"
              />
            </div>
          </div>

          <div class="flex justify-end pt-1">
            <button
              type="button"
              @click="handleResetPasswordByOTP"
              :disabled="isResettingPwdOTP || !otpResetCode || otpResetCode.length !== 6 || !otpResetNewPassword"
              class="px-5 py-2.5 bg-status-up hover:bg-status-up text-black font-bold text-xs font-mono rounded-xl shadow-md shadow-status-up/20 disabled:opacity-50 flex items-center gap-1.5 cursor-pointer transition-colors"
            >
              <Check class="w-4 h-4" :class="isResettingPwdOTP ? 'animate-spin' : ''" />
              <span>{{ isResettingPwdOTP ? 'Saving...' : 'Update Password via Email' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Section 4: Multi-Factor Authentication (MFA / 2FA) -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-bold text-text-main uppercase tracking-wider font-mono flex items-center gap-2">
            <ShieldCheck class="w-4 h-4 text-brand-periwinkle" />
            Two-Factor Authentication (MFA / 2FA)
          </h2>
          <span
            class="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase border"
            :class="mfaEnabled ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30' : 'bg-gray-500/10 text-text-secondary border-gray-500/30'"
          >
            {{ mfaEnabled ? 'MFA ACTIVE' : 'MFA DISABLED' }}
          </span>
        </div>

        <p class="text-xs text-text-secondary font-mono">
          Protect your SANOC account using Google Authenticator or Authy. Requiring a 6-digit TOTP code during login prevents unauthorized access.
        </p>

        <div class="flex items-center gap-3">
          <button
            v-if="!mfaEnabled"
            type="button"
            @click="handleInitiateMFASetup"
            class="px-4 py-2 bg-brand-periwinkle/10 hover:bg-brand-periwinkle/20 text-brand-periwinkle border border-brand-periwinkle/30 rounded-lg text-xs font-mono font-bold flex items-center gap-2 transition-colors cursor-pointer"
          >
            <ShieldCheck class="w-4 h-4" />
            <span>Enable 2FA / MFA</span>
          </button>
          <button
            v-else
            type="button"
            @click="handleDisableMFA"
            class="px-4 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 rounded-lg text-xs font-mono font-bold flex items-center gap-2 transition-colors cursor-pointer"
          >
            <ShieldOff class="w-4 h-4" />
            <span>Disable 2FA / MFA</span>
          </button>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="pt-4 border-t border-subtle flex items-center justify-end gap-3">
        <button
          type="button"
          @click="resetForm"
          class="px-4 py-2 bg-card hover:bg-subtle text-text-secondary rounded-lg text-xs font-mono transition-colors"
        >
          Discard Changes
        </button>
        <button
          type="button"
          @click="saveProfile"
          :disabled="isSubmitting"
          class="px-5 py-2 bg-brand-periwinkle hover:bg-brand-periwinkle-hover text-white rounded-lg text-xs font-mono font-bold flex items-center gap-2 shadow-lg shadow-brand-periwinkle/20 disabled:opacity-50 transition-colors"
        >
          <Save v-if="!isSubmitting" class="w-4 h-4" />
          <RefreshCw v-else class="w-4 h-4 animate-spin" />
          <span>{{ isSubmitting ? 'Saving Profile...' : 'Save Profile Changes' }}</span>
        </button>
      </div>
    </div>

    <!-- MFA Setup Modal -->
    <div v-if="showMFASetupModal" class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50 animate-fadeIn">
      <div class="bg-surface border border-brand-periwinkle/30 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-5">
        <div class="flex items-center justify-between border-b border-subtle pb-3">
          <h3 class="text-sm font-bold text-text-main font-mono flex items-center gap-2">
            <ShieldCheck class="w-4.5 h-4.5 text-brand-periwinkle" />
            Set Up Two-Factor Authentication
          </h3>
          <button @click="showMFASetupModal = false" class="text-text-secondary hover:text-text-main">
            <X class="w-4 h-4" />
          </button>
        </div>

        <div class="space-y-4">
          <p class="text-xs text-text-secondary font-mono text-center">
            Scan this QR code using <strong>Google Authenticator</strong>, <strong>Authy</strong>, or your password manager:
          </p>

          <!-- QR Code Display Box -->
          <div class="flex flex-col items-center justify-center p-4 bg-white rounded-2xl border border-subtle shadow-inner mx-auto max-w-[210px]">
            <img v-if="qrCodeDataUrl" :src="qrCodeDataUrl" alt="2FA QR Code" class="w-40 h-40 object-contain rounded" />
            <div v-else class="w-40 h-40 flex items-center justify-center text-xs text-text-muted font-mono">
              Generating QR...
            </div>
          </div>

          <!-- Secret Key Fallback Box -->
          <div class="bg-card border border-subtle p-3 rounded-xl text-center space-y-1">
            <span class="text-[10px] font-mono uppercase text-text-secondary">Or enter manual key</span>
            <div class="text-xs font-mono font-bold text-brand-periwinkle tracking-wider select-all">{{ mfaSecret }}</div>
          </div>

          <!-- 6-digit confirmation code input -->
          <div class="space-y-2 pt-1">
            <label class="block text-center text-[10px] font-mono uppercase tracking-wider text-text-secondary font-semibold">
              Enter 6-Digit Passcode from App
            </label>
            <OtpInput
              v-model="mfaSetupCode"
              :error="!!mfaSetupError"
              :disabled="isSubmittingMFA"
              @complete="handleConfirmMFASetup"
            />
          </div>

          <div v-if="mfaSetupError" class="text-xs text-red-400 font-mono flex items-center justify-center gap-1.5">
            <AlertTriangle class="w-3.5 h-3.5 shrink-0" />
            <span>{{ mfaSetupError }}</span>
          </div>

          <div class="flex items-center gap-3 pt-2">
            <button
              type="button"
              @click="showMFASetupModal = false"
              class="flex-1 py-2.5 bg-subtle hover:bg-hover text-text-secondary rounded-xl text-xs font-mono font-bold transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="button"
              @click="handleConfirmMFASetup"
              :disabled="mfaSetupCode.length !== 6 || isSubmittingMFA"
              class="flex-1 py-2.5 bg-brand-periwinkle hover:bg-brand-periwinkle-hover disabled:opacity-50 text-white rounded-xl text-xs font-mono font-bold transition-colors flex items-center justify-center gap-2 cursor-pointer"
            >
              <span>{{ isSubmittingMFA ? 'Verifying...' : 'Activate 2FA' }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useAuthStore } from '../stores/authStore';
import OtpInput from '../components/common/OtpInput.vue';
import QRCode from 'qrcode';
import { useSettingStore } from '../stores/settingStore';
import {
  User,
  Camera,
  Mail,
  Lock,
  KeyRound,
  Check,
  CheckCircle,
  AlertTriangle,
  X,
  Save,
  RefreshCw,
  Send,
  ShieldCheck,
  ShieldOff
} from 'lucide-vue-next';
import { authApi } from '../api';

const authStore = useAuthStore();
const settingStore = useSettingStore();
const fileInputRef = ref<HTMLInputElement | null>(null);

const isSubmitting = ref(false);
const toastMessage = ref('');
const toastType = ref<'success' | 'error'>('success');

// Password change method selection
const pwdChangeMethod = ref<'current_password' | 'email_otp'>('current_password');

// OTP Reset Password states
const otpResetCode = ref('');
const otpResetNewPassword = ref('');
const otpResetConfirmPassword = ref('');
const isSendingProfileOTP = ref(false);
const isResettingPwdOTP = ref(false);
const profileOTPCountdown = ref(0);
let profileOTPTimer: any = null;

function startProfileOTPCountdown() {
  profileOTPCountdown.value = 60;
  if (profileOTPTimer) clearInterval(profileOTPTimer);
  profileOTPTimer = setInterval(() => {
    if (profileOTPCountdown.value > 0) {
      profileOTPCountdown.value--;
    } else {
      clearInterval(profileOTPTimer);
      profileOTPTimer = null;
    }
  }, 1000);
}

async function handleSendProfileResetOTP() {
  if (!form.email) {
    toastType.value = 'error';
    toastMessage.value = 'Email akun tidak ditemukan.';
    return;
  }
  isSendingProfileOTP.value = true;
  try {
    const res = await authApi.sendProfileResetOTP(form.email);
    startProfileOTPCountdown();
    toastType.value = 'success';
    toastMessage.value = res.message || `Kode OTP verifikasi telah dikirimkan ke ${form.email}`;
  } catch (e: any) {
    toastType.value = 'error';
    toastMessage.value = e.response?.data?.error || 'Gagal mengirimkan kode OTP ke email.';
  } finally {
    isSendingProfileOTP.value = false;
  }
}

async function handleResetPasswordByOTP() {
  if (!otpResetCode.value || otpResetCode.value.trim().length !== 6) {
    toastType.value = 'error';
    toastMessage.value = 'Masukkan 6-digit kode verifikasi OTP email.';
    return;
  }
  if (!otpResetNewPassword.value || otpResetNewPassword.value.length < 8) {
    toastType.value = 'error';
    toastMessage.value = 'Kata sandi baru minimal 8 karakter.';
    return;
  }
  if (otpResetNewPassword.value !== otpResetConfirmPassword.value) {
    toastType.value = 'error';
    toastMessage.value = 'Konfirmasi kata sandi tidak cocok dengan kata sandi baru.';
    return;
  }

  isResettingPwdOTP.value = true;
  try {
    const res = await authApi.resetProfilePasswordOTP({
      email: form.email,
      code: otpResetCode.value.trim(),
      newPassword: otpResetNewPassword.value
    });
    toastType.value = 'success';
    toastMessage.value = res.message || 'Kata sandi akun Anda berhasil diperbarui via verifikasi email!';
    otpResetCode.value = '';
    otpResetNewPassword.value = '';
    otpResetConfirmPassword.value = '';
    if (profileOTPTimer) clearInterval(profileOTPTimer);
    profileOTPCountdown.value = 0;
  } catch (e: any) {
    toastType.value = 'error';
    toastMessage.value = e.response?.data?.error || 'Gagal mereset kata sandi via email OTP.';
  } finally {
    isResettingPwdOTP.value = false;
  }
}

const mfaEnabled = ref(false);
const showMFASetupModal = ref(false);
const mfaSecret = ref('');
const mfaOtpUri = ref('');
const mfaSetupCode = ref('');
const mfaSetupError = ref('');
const qrCodeDataUrl = ref('');
const isSubmittingMFA = ref(false);

const presetAvatars = [
  'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256',
  'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=256',
  'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=256',
  'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&q=80&w=256',
  'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?auto=format&fit=crop&q=80&w=256'
];

const form = reactive({
  username: '',
  name: '',
  email: '',
  avatarUrl: '',
  currentPassword: '',
  newPassword: ''
});

const passwordLengthValid = computed(() => form.newPassword.length >= 12);

const passwordScore = computed(() => {
  const p = form.newPassword;
  if (!p) return 0;
  let score = 0;
  if (p.length >= 12) score++;
  if (/[A-Z]/.test(p)) score++;
  if (/[0-9]/.test(p)) score++;
  if (/[^A-Za-z0-9]/.test(p)) score++;
  return score;
});

const passwordStrengthText = computed(() => {
  switch (passwordScore.value) {
    case 0: return 'Too Short';
    case 1: return 'Weak';
    case 2: return 'Medium';
    case 3: return 'Strong';
    case 4: return 'Very Strong';
    default: return 'Weak';
  }
});

function loadProfile() {
  const u = authStore.user as any;
  const fallbackUsername = u.email ? u.email.split('@')[0] : '';
  form.username = u.username || fallbackUsername || '';
  form.name = u.name || '';
  form.email = u.email || '';
  form.avatarUrl = u.avatarUrl || presetAvatars[0];
  form.currentPassword = '';
  form.newPassword = '';
  mfaEnabled.value = u.mfaEnabled || false;
}

watch(
  () => authStore.user,
  () => {
    loadProfile();
  },
  { deep: true, immediate: true }
);

onMounted(() => {
  loadProfile();
});

function resetForm() {
  loadProfile();
  toastMessage.value = '';
}

function triggerFilePicker() {
  fileInputRef.value?.click();
}

async function handleInitiateMFASetup() {
  mfaSetupError.value = '';
  mfaSetupCode.value = '';
  qrCodeDataUrl.value = '';
  try {
    const res = await authStore.setupMFA();
    if (res.success && res.secret) {
      mfaSecret.value = res.secret;
      mfaOtpUri.value = res.otpAuthUri;
      const uri = res.otpAuthUri || `otpauth://totp/SANOC%20Monitoring:${encodeURIComponent(authStore.user?.email || 'user')}?secret=${res.secret}&issuer=SANOC%20Monitoring`;
      qrCodeDataUrl.value = await QRCode.toDataURL(uri, {
        width: 200,
        margin: 1,
        color: {
          dark: '#000000',
          light: '#ffffff'
        }
      });
      showMFASetupModal.value = true;
    } else {
      toastType.value = 'error';
      toastMessage.value = res.message || 'Failed to initiate MFA setup.';
    }
  } catch (e: any) {
    toastType.value = 'error';
    toastMessage.value = e?.message || 'Failed to initiate MFA setup.';
  }
}

async function handleConfirmMFASetup() {
  mfaSetupError.value = '';
  isSubmittingMFA.value = true;

  try {
    const res = await authStore.verifyMFA(mfaSecret.value, mfaSetupCode.value);
    if (res.success) {
      mfaEnabled.value = true;
      (authStore.user as any).mfaEnabled = true;
      showMFASetupModal.value = false;
      toastType.value = 'success';
      toastMessage.value = 'Two-Factor Authentication (MFA) has been enabled!';
    } else {
      mfaSetupError.value = res.message || 'Invalid 6-digit passcode';
    }
  } catch (e: any) {
    mfaSetupError.value = e?.message || 'Invalid 6-digit passcode';
  } finally {
    isSubmittingMFA.value = false;
  }
}

async function handleDisableMFA() {
  try {
    const res = await authStore.disableMFA();
    if (res.success) {
      mfaEnabled.value = false;
      (authStore.user as any).mfaEnabled = false;
      toastType.value = 'success';
      toastMessage.value = 'Two-Factor Authentication (MFA) has been disabled.';
    } else {
      toastType.value = 'error';
      toastMessage.value = res.message || 'Failed to disable MFA.';
    }
  } catch (e: any) {
    toastType.value = 'error';
    toastMessage.value = e?.message || 'Failed to disable MFA.';
  }
}

async function handleFileUpload(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;

  if (file.size > 5 * 1024 * 1024) {
    toastType.value = 'error';
    toastMessage.value = 'File size exceeds maximum 5MB limit.';
    return;
  }

  isSubmitting.value = true;
  toastMessage.value = 'Uploading image...';
  toastType.value = 'success';

  try {
    const res = await authStore.uploadAvatar(file);
    if (res.success && res.avatarUrl) {
      form.avatarUrl = res.avatarUrl;
      toastType.value = 'success';
      toastMessage.value = 'Profile picture uploaded and saved successfully!';
      await settingStore.fetchUsers();
      return;
    }
  } catch (err: any) {
    // fallback
  } finally {
    isSubmitting.value = false;
  }

  const reader = new FileReader();
  reader.onload = (e) => {
    if (e.target?.result) {
      const rawData = e.target.result as string;
      form.avatarUrl = rawData;
      toastType.value = 'success';
      toastMessage.value = 'Local photo selected! Click "Save Profile Changes" to update.';
    }
  };
  reader.readAsDataURL(file);
}

async function saveProfile() {
  if (!form.name.trim() || !form.email.trim()) {
    toastType.value = 'error';
    toastMessage.value = 'Name and email address are required.';
    return;
  }

  if (form.newPassword && form.newPassword.length < 12) {
    toastType.value = 'error';
    toastMessage.value = 'New password must be at least 12 characters.';
    return;
  }

  isSubmitting.value = true;
  toastMessage.value = '';

  try {
    const res = await authStore.updateProfile({
      username: form.username.trim(),
      name: form.name.trim(),
      email: form.email.trim(),
      avatarUrl: form.avatarUrl.trim(),
      currentPassword: form.currentPassword,
      newPassword: form.newPassword
    });

    if (res.success) {
      toastType.value = 'success';
      toastMessage.value = res.message || 'User profile updated successfully!';
      form.currentPassword = '';
      form.newPassword = '';
      await settingStore.fetchUsers();
    } else {
      toastType.value = 'error';
      toastMessage.value = res.message || 'Failed to update profile details.';
    }
  } catch (err: any) {
    toastType.value = 'error';
    toastMessage.value = err?.message || 'An error occurred while saving profile.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>
