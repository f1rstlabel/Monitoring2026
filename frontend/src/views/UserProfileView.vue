<template>
  <div class="space-y-6 max-w-4xl mx-auto pb-12">
    <!-- Header Page Banner -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-100 flex items-center gap-2.5">
          <User class="w-6 h-6 text-[#7B96F5]" />
          User Profile &amp; Account Settings
        </h1>
        <p class="text-xs text-gray-400 mt-1">
          Manage your personal profile details, profile picture, email address, and security password.
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
    <div class="bg-[#151517] border border-[#26262A] rounded-xl p-6 space-y-6 shadow-xl">
      <!-- Section 1: Profile Picture & Avatar -->
      <div class="border-b border-[#26262A] pb-6 space-y-4">
        <h2 class="text-sm font-bold text-gray-200 uppercase tracking-wider font-mono flex items-center gap-2">
          <Camera class="w-4 h-4 text-[#7B96F5]" />
          Profile Avatar &amp; Picture
        </h2>

        <div class="flex flex-col sm:flex-row items-center gap-6">
          <!-- Current Avatar Preview with Change Photo Overlay -->
          <div class="relative group cursor-pointer shrink-0" @click="triggerFilePicker" title="Click to upload new profile photo">
            <img
              :src="form.avatarUrl || 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'"
              alt="Profile Avatar"
              @error="(e: Event) => (e.target as HTMLImageElement).src = 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256'"
              class="w-24 h-24 rounded-full object-cover border-2 border-[#7B96F5]/50 shadow-lg ring-4 ring-[#151517] group-hover:opacity-80 transition-all"
            />
            <div class="absolute inset-0 bg-black/60 rounded-full flex flex-col items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity text-[10px] font-mono font-bold">
              <Camera class="w-5 h-5 mb-0.5" />
              <span>Change Photo</span>
            </div>
            <!-- Camera badge icon at bottom right -->
            <div class="absolute bottom-0 right-0 p-1.5 bg-[#7B96F5] text-white rounded-full shadow-md border-2 border-[#151517]">
              <Camera class="w-3.5 h-3.5" />
            </div>
          </div>

          <!-- Direct Change Photo Action & Helper Text -->
          <div class="flex-1 space-y-2 w-full text-center sm:text-left">
            <div>
              <h3 class="text-xs font-bold text-gray-200 font-mono">Profile Photo</h3>
              <p class="text-[11px] text-gray-400 font-mono mt-0.5">
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
      <div class="border-b border-[#26262A] pb-6 space-y-4">
        <h2 class="text-sm font-bold text-gray-200 uppercase tracking-wider font-mono flex items-center gap-2">
          <Mail class="w-4 h-4 text-[#3ECF8E]" />
          Account Information
        </h2>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1 font-mono">Username *</label>
            <input
              v-model="form.username"
              type="text"
              required
              placeholder="e.g. admin.noc"
              class="w-full bg-[#1A1A1E] border border-[#26262A] focus:border-[#3ECF8E] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1 font-mono">Full Name *</label>
            <input
              v-model="form.name"
              type="text"
              required
              placeholder="e.g. Budi Santoso"
              class="w-full bg-[#1A1A1E] border border-[#26262A] focus:border-[#3ECF8E] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1 font-mono">Email Address *</label>
            <input
              v-model="form.email"
              type="email"
              required
              placeholder="e.g. user@jabarprov.go.id"
              class="w-full bg-[#1A1A1E] border border-[#26262A] focus:border-[#3ECF8E] rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none font-mono"
            />
          </div>
        </div>
      </div>

      <!-- Section 3: Password Update (Minimum 12 Characters) -->
      <div class="space-y-4 border-b border-[#26262A] pb-6">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-bold text-gray-200 uppercase tracking-wider font-mono flex items-center gap-2">
            <Lock class="w-4 h-4 text-amber-400" />
            Security &amp; Change Password
          </h2>
          <span class="text-[10px] font-mono text-gray-500">Leave blank to keep current password</span>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1 font-mono">Current Password</label>
            <input
              v-model="form.currentPassword"
              type="password"
              placeholder="Enter current password"
              class="w-full bg-[#1A1A1E] border border-[#26262A] focus:border-amber-400 rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-gray-400 mb-1 font-mono">New Password (Min 12 Chars)</label>
            <input
              v-model="form.newPassword"
              type="password"
              placeholder="Minimum 12 characters"
              class="w-full bg-[#1A1A1E] border border-[#26262A] focus:border-amber-400 rounded-lg px-3 py-2 text-xs text-gray-200 focus:outline-none font-mono"
            />
            <div v-if="form.newPassword" class="mt-2 space-y-1">
              <div class="flex items-center justify-between text-[10px] font-mono">
                <span class="text-gray-400">Strength: {{ passwordStrengthText }}</span>
                <span :class="passwordLengthValid ? 'text-emerald-400' : 'text-red-400'">{{ form.newPassword.length }}/12</span>
              </div>
              <div class="w-full h-1 bg-[#26262A] rounded-full overflow-hidden">
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
      </div>

      <!-- Section 4: Multi-Factor Authentication (MFA / 2FA) -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-bold text-gray-200 uppercase tracking-wider font-mono flex items-center gap-2">
            <ShieldCheck class="w-4 h-4 text-[#7B96F5]" />
            Two-Factor Authentication (MFA / 2FA)
          </h2>
          <span
            class="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase border"
            :class="mfaEnabled ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30' : 'bg-gray-500/10 text-gray-400 border-gray-500/30'"
          >
            {{ mfaEnabled ? 'MFA ACTIVE' : 'MFA DISABLED' }}
          </span>
        </div>

        <p class="text-xs text-gray-400 font-mono">
          Protect your SANOC account using Google Authenticator or Authy. Requiring a 6-digit TOTP code during login prevents unauthorized access.
        </p>

        <div class="flex items-center gap-3">
          <button
            v-if="!mfaEnabled"
            type="button"
            @click="handleInitiateMFASetup"
            class="px-4 py-2 bg-[#7B96F5]/10 hover:bg-[#7B96F5]/20 text-[#7B96F5] border border-[#7B96F5]/30 rounded-lg text-xs font-mono font-bold flex items-center gap-2 transition-colors cursor-pointer"
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
      <div class="pt-4 border-t border-[#26262A] flex items-center justify-end gap-3">
        <button
          type="button"
          @click="resetForm"
          class="px-4 py-2 bg-[#1C1C1F] hover:bg-[#26262A] text-gray-400 rounded-lg text-xs font-mono transition-colors"
        >
          Discard Changes
        </button>
        <button
          type="button"
          @click="saveProfile"
          :disabled="isSubmitting"
          class="px-5 py-2 bg-[#7B96F5] hover:bg-[#6884E6] text-white rounded-lg text-xs font-mono font-bold flex items-center gap-2 shadow-lg shadow-[#7B96F5]/20 disabled:opacity-50 transition-colors"
        >
          <Save v-if="!isSubmitting" class="w-4 h-4" />
          <RefreshCw v-else class="w-4 h-4 animate-spin" />
          <span>{{ isSubmitting ? 'Saving Profile...' : 'Save Profile Changes' }}</span>
        </button>
      </div>
    </div>

    <!-- MFA Setup Modal -->
    <div v-if="showMFASetupModal" class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50 animate-fadeIn">
      <div class="bg-[#151517] border border-[#7B96F5]/30 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-5">
        <div class="flex items-center justify-between border-b border-[#26262A] pb-3">
          <h3 class="text-sm font-bold text-white font-mono flex items-center gap-2">
            <ShieldCheck class="w-4.5 h-4.5 text-[#7B96F5]" />
            Set Up Two-Factor Authentication
          </h3>
          <button @click="showMFASetupModal = false" class="text-gray-400 hover:text-white">
            <X class="w-4 h-4" />
          </button>
        </div>

        <div class="space-y-4">
          <p class="text-xs text-gray-300 font-mono">
            Scan the QR code below or enter the Secret Key manually into Google Authenticator or Authy.
          </p>

          <!-- Secret Key Box -->
          <div class="bg-[#18181B] border border-[#26262A] p-3 rounded-lg text-center space-y-1">
            <span class="text-[10px] font-mono uppercase text-gray-400">Secret Key</span>
            <div class="text-sm font-mono font-bold text-[#7B96F5] tracking-wider select-all">{{ mfaSecret }}</div>
          </div>

          <!-- 6-digit confirmation code input -->
          <div class="space-y-1.5 pt-2">
            <label class="block text-[10px] font-mono uppercase tracking-wider text-gray-400 font-semibold">Enter 6-Digit Passcode from App</label>
            <input
              v-model="mfaSetupCode"
              type="text"
              maxlength="6"
              placeholder="123456"
              class="w-full bg-[#18181B] border border-[#26262A] focus:border-[#7B96F5] rounded-lg px-4 py-2.5 text-center text-base tracking-[0.4em] font-mono text-white placeholder-gray-600 focus:outline-none transition-colors"
            />
          </div>

          <div v-if="mfaSetupError" class="text-xs text-red-400 font-mono flex items-center gap-1.5">
            <AlertTriangle class="w-3.5 h-3.5 shrink-0" />
            <span>{{ mfaSetupError }}</span>
          </div>

          <div class="flex items-center gap-3 pt-2">
            <button
              type="button"
              @click="showMFASetupModal = false"
              class="flex-1 py-2.5 bg-[#26262A] hover:bg-[#333338] text-gray-300 rounded-lg text-xs font-mono font-bold transition-colors"
            >
              Cancel
            </button>
            <button
              type="button"
              @click="handleConfirmMFASetup"
              :disabled="mfaSetupCode.length !== 6 || isSubmittingMFA"
              class="flex-1 py-2.5 bg-[#7B96F5] hover:bg-[#95ABF7] disabled:opacity-50 text-white rounded-lg text-xs font-mono font-bold transition-colors flex items-center justify-center gap-2"
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
import { useSettingStore } from '../stores/settingStore';
import {
  User,
  Camera,
  Mail,
  Lock,
  CheckCircle,
  AlertTriangle,
  X,
  Save,
  RefreshCw,
  ShieldCheck,
  ShieldOff
} from 'lucide-vue-next';

const authStore = useAuthStore();
const settingStore = useSettingStore();
const fileInputRef = ref<HTMLInputElement | null>(null);

const isSubmitting = ref(false);
const toastMessage = ref('');
const toastType = ref<'success' | 'error'>('success');

const mfaEnabled = ref(false);
const showMFASetupModal = ref(false);
const mfaSecret = ref('');
const mfaOtpUri = ref('');
const mfaSetupCode = ref('');
const mfaSetupError = ref('');
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
  try {
    const res = await authStore.setupMFA();
    if (res.success && res.secret) {
      mfaSecret.value = res.secret;
      mfaOtpUri.value = res.otpAuthUri;
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
