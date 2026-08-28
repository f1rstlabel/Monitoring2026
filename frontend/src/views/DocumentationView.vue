<template>
  <div class="space-y-6 max-w-7xl mx-auto pb-16 font-sans text-text-main">
    <!-- Header Hero Banner -->
    <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-surface via-card to-main border border-subtle p-6 sm:p-8 shadow-2xl">
      <div class="absolute -right-16 -top-16 w-80 h-80 bg-brand-periwinkle/10 rounded-full blur-3xl pointer-events-none"></div>
      
      <div class="relative z-10 flex flex-col lg:flex-row lg:items-center justify-between gap-6">
        <div class="space-y-3 max-w-2xl">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full text-[11px] font-mono font-semibold uppercase bg-brand-periwinkle/15 text-brand-periwinkle border border-brand-periwinkle/30">
            <HelpCircle class="w-3.5 h-3.5" />
            <span>{{ t.badge }}</span>
          </div>
          
          <h1 class="text-2xl sm:text-3xl font-extrabold text-text-main tracking-tight leading-tight">
            {{ t.title }}
          </h1>
          
          <p class="text-xs sm:text-sm text-text-secondary leading-relaxed font-sans">
            {{ t.subtitle }}
          </p>
        </div>

        <!-- Search Bar & Dedicated Language Switcher -->
        <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 w-full lg:w-auto shrink-0">
          <div class="relative w-full sm:w-72">
            <Search class="w-4 h-4 text-text-secondary absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t.searchPlaceholder"
              class="w-full bg-card border border-subtle rounded-xl pl-10 pr-9 py-2.5 text-xs text-text-main focus:outline-none focus:border-brand-periwinkle placeholder-text-muted font-mono shadow-inner transition-colors"
            />
            <button
              v-if="searchQuery"
              @click="searchQuery = ''"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary cursor-pointer p-0.5 rounded"
              title="Clear search"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>

          <!-- Pusat Bantuan Language Switcher -->
          <div class="flex items-center self-end sm:self-auto bg-card border border-subtle rounded-xl p-1 shadow-sm shrink-0">
            <button
              @click="langStore.setLang('id')"
              class="px-2.5 py-1.5 rounded-lg text-xs font-mono font-bold transition-all flex items-center gap-1.5 cursor-pointer"
              :class="langStore.currentLang === 'id' ? 'bg-brand-periwinkle text-white shadow-sm' : 'text-text-secondary hover:text-text-main'"
              title="Bahasa Indonesia"
            >
              <span>🇮🇩</span>
              <span>ID</span>
            </button>
            <button
              @click="langStore.setLang('en')"
              class="px-2.5 py-1.5 rounded-lg text-xs font-mono font-bold transition-all flex items-center gap-1.5 cursor-pointer"
              :class="langStore.currentLang === 'en' ? 'bg-brand-periwinkle text-white shadow-sm' : 'text-text-secondary hover:text-text-main'"
              title="English"
            >
              <span>🇬🇧</span>
              <span>EN</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Layout: 2-Column Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
      <!-- Left Sub-Sidebar Navigation -->
      <aside class="lg:col-span-1 bg-surface border border-subtle rounded-2xl p-3 space-y-1.5 shadow-xl sticky top-20">
        <div class="px-3 py-2 border-b border-subtle/60 mb-2">
          <span class="text-[10px] font-mono font-bold uppercase tracking-wider text-text-secondary">{{ t.navHeader }}</span>
        </div>

        <nav class="space-y-1" aria-label="Documentation Sections">
          <button
            v-for="cat in categories"
            :key="cat.id"
            @click="activeSection = cat.id"
            class="w-full text-left p-3 rounded-xl transition-all flex items-start gap-3 group relative cursor-pointer"
            :class="[
              activeSection === cat.id
                ? 'bg-brand-periwinkle/10 border border-brand-periwinkle/30 text-text-main shadow-sm shadow-brand-periwinkle/10'
                : 'border border-transparent text-text-secondary hover:text-text-main hover:bg-card hover:border-subtle'
            ]"
          >
            <component
              :is="cat.icon"
              class="w-4 h-4 shrink-0 mt-0.5 transition-colors"
              :class="activeSection === cat.id ? 'text-brand-periwinkle' : 'text-text-secondary group-hover:text-text-main'"
            />
            <div class="flex-1 min-w-0">
              <span class="text-xs font-bold font-mono tracking-tight block truncate" :class="activeSection === cat.id ? 'text-text-main' : 'text-text-secondary'">
                {{ cat.label[lang] }}
              </span>
              <p class="text-[10px] text-text-muted truncate mt-0.5 font-sans">{{ cat.description[lang] }}</p>
            </div>
            <span
              v-if="activeSection === cat.id"
              class="w-1.5 h-1.5 rounded-full bg-brand-periwinkle absolute right-2.5 top-1/2 -translate-y-1/2"
            ></span>
          </button>
        </nav>

        <!-- Quick Helpdesk Widget -->
        <div class="mt-4 p-3.5 bg-card border border-subtle rounded-xl space-y-2 text-xs">
          <div class="flex items-center gap-2 text-text-main font-mono font-bold text-[11px]">
            <Headphones class="w-3.5 h-3.5 text-brand-periwinkle" />
            <span>{{ t.quickHelpTitle }}</span>
          </div>
          <p class="text-[10px] text-text-secondary leading-relaxed font-sans">
            {{ t.quickHelpDesc }}
          </p>
          <div class="pt-2 border-t border-subtle text-[10px] font-mono text-brand-periwinkle space-y-0.5">
            <div class="select-all font-semibold">noc.alerts@jabarprov.go.id</div>
            <div class="text-text-secondary">Ext: 4401 / 4402 (Hotline)</div>
          </div>
        </div>
      </aside>

      <!-- Right Main Content -->
      <main class="lg:col-span-3 space-y-6 min-w-0">
        
        <!-- ══════════════════════════════════════════════════════════════════════
             SECTION 1: User Guides (Panduan Pengguna)
             ══════════════════════════════════════════════════════════════════════ -->
        <section v-if="activeSection === 'guides'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <BookOpen class="w-4 h-4 text-brand-periwinkle" />
                {{ t.guidesHeader }}
              </h2>
              <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.guidesSubheader }}</p>
            </div>
          </div>

          <!-- Guide Card 1: Dashboard Monitoring -->
          <div class="bg-surface border border-subtle rounded-2xl p-6 space-y-4 shadow-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-brand-periwinkle/15 border border-brand-periwinkle/30 flex items-center justify-center text-brand-periwinkle">
                <LayoutGrid class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-sm font-bold text-text-main font-mono">1. {{ t.guide1Title }}</h3>
                <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.guide1Sub }}</p>
              </div>
            </div>

            <div class="space-y-3 text-xs text-text-secondary leading-relaxed font-sans pt-2 border-t border-subtle">
              <p>{{ t.guide1Desc }}</p>
              <ul class="list-disc list-inside space-y-1.5 pl-2 text-text-secondary font-sans text-xs">
                <li><strong class="text-text-main font-mono">Summary Metrics:</strong> {{ t.guide1Bullet1 }}</li>
                <li><strong class="text-text-main font-mono">Live Activity Feed:</strong> {{ t.guide1Bullet2 }}</li>
                <li><strong class="text-text-main font-mono">Top Flapping Devices:</strong> {{ t.guide1Bullet3 }}</li>
                <li><strong class="text-text-main font-mono">Refresh Now:</strong> {{ t.guide1Bullet4 }}</li>
              </ul>
            </div>
          </div>

          <!-- Guide Card 2: Devices Management & Bulk Mode -->
          <div class="bg-surface border border-subtle rounded-2xl p-6 space-y-4 shadow-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
                <Server class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-sm font-bold text-text-main font-mono">2. {{ t.guide2Title }}</h3>
                <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.guide2Sub }}</p>
              </div>
            </div>

            <div class="space-y-3 text-xs text-text-secondary leading-relaxed font-sans pt-2 border-t border-subtle">
              <p>{{ t.guide2Desc }}</p>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
                <div class="p-4 bg-card border border-subtle rounded-xl space-y-1.5">
                  <span class="text-[10px] font-mono font-bold text-brand-periwinkle uppercase tracking-wider">{{ t.guide2Card1Title }}</span>
                  <p class="text-xs text-text-secondary font-sans leading-relaxed">{{ t.guide2Card1Desc }}</p>
                </div>
                <div class="p-4 bg-card border border-subtle rounded-xl space-y-1.5">
                  <span class="text-[10px] font-mono font-bold text-emerald-400 uppercase tracking-wider">{{ t.guide2Card2Title }}</span>
                  <p class="text-xs text-text-secondary font-sans leading-relaxed">{{ t.guide2Card2Desc }}</p>
                </div>
                <div class="p-4 bg-card border border-subtle rounded-xl space-y-1.5">
                  <span class="text-[10px] font-mono font-bold text-amber-400 uppercase tracking-wider">{{ t.guide2Card3Title }}</span>
                  <p class="text-xs text-text-secondary font-sans leading-relaxed">{{ t.guide2Card3Desc }}</p>
                </div>
                <div class="p-4 bg-card border border-subtle rounded-xl space-y-1.5">
                  <span class="text-[10px] font-mono font-bold text-sky-400 uppercase tracking-wider">{{ t.guide2Card4Title }}</span>
                  <p class="text-xs text-text-secondary font-sans leading-relaxed">{{ t.guide2Card4Desc }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Guide Card 3: Incidents & Flap Reuse -->
          <div class="bg-surface border border-subtle rounded-2xl p-6 space-y-4 shadow-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-red-500/15 border border-red-500/30 flex items-center justify-center text-red-400">
                <AlertTriangle class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-sm font-bold text-text-main font-mono">3. {{ t.guide3Title }}</h3>
                <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.guide3Sub }}</p>
              </div>
            </div>

            <div class="space-y-3 text-xs text-text-secondary leading-relaxed font-sans pt-2 border-t border-subtle">
              <p>{{ t.guide3Desc }}</p>
              <ul class="list-disc list-inside space-y-1.5 pl-2 text-text-secondary font-sans text-xs">
                <li><strong class="text-text-main font-mono">DOWN Confirmation:</strong> {{ t.guide3Bullet1 }}</li>
                <li><strong class="text-text-main font-mono">Auto-Resolution (UP):</strong> {{ t.guide3Bullet2 }}</li>
                <li><strong class="text-text-main font-mono">Flap Reuse Window:</strong> {{ t.guide3Bullet3 }}</li>
                <li><strong class="text-text-main font-mono">Audit Log &amp; Fallback:</strong> {{ t.guide3Bullet4 }}</li>
              </ul>
            </div>
          </div>

          <!-- Guide Card 4: Gateways & 2FA Security -->
          <div class="bg-surface border border-subtle rounded-2xl p-6 space-y-4 shadow-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-400">
                <Sliders class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-sm font-bold text-text-main font-mono">4. {{ t.guide4Title }}</h3>
                <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.guide4Sub }}</p>
              </div>
            </div>

            <div class="space-y-3 text-xs text-text-secondary leading-relaxed font-sans pt-2 border-t border-subtle">
              <p>{{ t.guide4Desc }}</p>
              <ul class="list-disc list-inside space-y-1.5 pl-2 text-text-secondary font-sans text-xs">
                <li><strong class="text-text-main font-mono">WhatsApp QR Gateway:</strong> {{ t.guide4Bullet1 }}</li>
                <li><strong class="text-text-main font-mono">Telegram Fallback Bot:</strong> {{ t.guide4Bullet2 }}</li>
                <li><strong class="text-text-main font-mono">Two-Factor Authentication:</strong> {{ t.guide4Bullet3 }}</li>
              </ul>
            </div>
          </div>
        </section>

        <!-- ══════════════════════════════════════════════════════════════════════
             SECTION 2: FAQ & QnA (Tanya Jawab Profesional & Rapi)
             ══════════════════════════════════════════════════════════════════════ -->
        <section v-else-if="activeSection === 'faq'" class="space-y-6 animate-fadeIn">
          <!-- Section Header -->
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <HelpCircle class="w-4 h-4 text-brand-periwinkle" />
                {{ t.faqHeader }}
              </h2>
              <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.faqSubheader }}</p>
            </div>
            <span class="text-xs font-mono text-brand-periwinkle font-semibold bg-brand-periwinkle/10 border border-brand-periwinkle/20 px-2.5 py-1 rounded-full self-start sm:self-auto">
              {{ filteredFaqs.length }} {{ t.faqCounterLabel }}
            </span>
          </div>

          <!-- FAQ Category Tag Filter Bar -->
          <div class="flex flex-wrap items-center gap-2">
            <button
              @click="selectedFaqTag = 'all'"
              class="px-3 py-1.5 rounded-lg text-xs font-mono font-semibold transition-all cursor-pointer border"
              :class="selectedFaqTag === 'all'
                ? 'bg-brand-periwinkle border-brand-periwinkle text-white shadow-md shadow-brand-periwinkle/20'
                : 'bg-surface border-subtle text-text-secondary hover:text-text-main hover:bg-card'"
            >
              {{ t.allCategories }}
            </button>
            <button
              v-for="tag in faqTags"
              :key="tag.key"
              @click="selectedFaqTag = tag.key"
              class="px-3 py-1.5 rounded-lg text-xs font-mono font-semibold transition-all cursor-pointer border"
              :class="selectedFaqTag === tag.key
                ? 'bg-brand-periwinkle border-brand-periwinkle text-white shadow-md shadow-brand-periwinkle/20'
                : 'bg-surface border-subtle text-text-secondary hover:text-text-main hover:bg-card'"
            >
              {{ tag.label[lang] }}
            </button>
          </div>

          <!-- Professional FAQ Cards Accordion -->
          <div class="space-y-3.5">
            <div
              v-for="(faq, idx) in filteredFaqs"
              :key="idx"
              class="bg-surface border rounded-2xl overflow-hidden transition-all shadow-xl"
              :class="openedFaqs.includes(idx) ? 'border-brand-periwinkle/40 bg-surface' : 'border-subtle hover:border-subtle'"
            >
              <!-- Question Header Row -->
              <button
                @click="toggleFaq(idx)"
                class="w-full p-5 text-left flex items-start justify-between gap-4 cursor-pointer hover:bg-card/70 transition-colors"
              >
                <div class="flex items-start gap-3.5 flex-1 min-w-0">
                  <span
                    class="w-7 h-7 rounded-xl flex items-center justify-center font-mono font-bold text-xs shrink-0 transition-colors mt-0.5 border"
                    :class="openedFaqs.includes(idx)
                      ? 'bg-brand-periwinkle border-brand-periwinkle text-white shadow-md shadow-brand-periwinkle/25'
                      : 'bg-card border-subtle text-brand-periwinkle'"
                  >
                    Q
                  </span>
                  <div class="space-y-1.5 flex-1">
                    <h3 class="text-xs sm:text-sm font-bold text-text-main font-sans leading-snug tracking-tight">
                      {{ faq.question[lang] }}
                    </h3>
                    <div class="flex items-center gap-2">
                      <span class="px-2 py-0.5 rounded text-[10px] font-mono font-medium bg-card border border-subtle text-text-secondary">
                        {{ faq.tag[lang] }}
                      </span>
                    </div>
                  </div>
                </div>

                <div
                  class="w-7 h-7 rounded-lg bg-card border border-subtle flex items-center justify-center text-text-secondary shrink-0 mt-0.5 transition-transform duration-200"
                  :class="openedFaqs.includes(idx) ? 'rotate-180 text-brand-periwinkle border-brand-periwinkle/30' : ''"
                >
                  <ChevronDown class="w-4 h-4" />
                </div>
              </button>

              <!-- Answer Body Panel -->
              <div
                v-if="openedFaqs.includes(idx)"
                class="px-5 pb-5 pt-1 border-t border-subtle/70 animate-fadeIn"
              >
                <div class="p-4 rounded-xl bg-main border border-subtle/60 flex items-start gap-3.5">
                  <span class="w-6 h-6 rounded-lg bg-emerald-500/15 border border-emerald-500/30 text-emerald-400 text-[11px] font-bold font-mono flex items-center justify-center shrink-0 mt-0.5">
                    A
                  </span>
                  <div class="text-xs text-text-secondary font-sans leading-relaxed space-y-2.5 flex-1">
                    <div v-html="faq.answer[lang]" class="faq-content"></div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Empty Search State -->
            <div v-if="filteredFaqs.length === 0" class="p-12 text-center bg-surface border border-subtle rounded-2xl space-y-3">
              <div class="w-12 h-12 rounded-2xl bg-card border border-subtle flex items-center justify-center text-text-muted mx-auto">
                <Search class="w-6 h-6" />
              </div>
              <h4 class="text-sm font-bold text-text-main font-mono">{{ t.noResultTitle }}</h4>
              <p class="text-xs text-text-muted max-w-sm mx-auto font-sans">{{ t.noResultDesc }}</p>
              <button
                @click="searchQuery = ''; selectedFaqTag = 'all'"
                class="px-4 py-2 rounded-xl bg-brand-periwinkle/10 border border-brand-periwinkle/30 text-brand-periwinkle text-xs font-mono font-semibold hover:bg-brand-periwinkle/20 cursor-pointer"
              >
                Reset Filter Pencarian
              </button>
            </div>
          </div>
        </section>

        <!-- ══════════════════════════════════════════════════════════════════════
             SECTION 3: Architecture & Topologies (Arsitektur Sistem)
             ══════════════════════════════════════════════════════════════════════ -->
        <section v-else-if="activeSection === 'architecture'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Network class="w-4 h-4 text-brand-periwinkle" />
                {{ t.archHeader }}
              </h2>
              <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.archSubheader }}</p>
            </div>
          </div>

          <div class="bg-surface border border-subtle rounded-2xl p-6 space-y-5 shadow-xl">
            <h3 class="text-xs font-bold text-text-main font-mono uppercase tracking-wider">{{ t.archFlowTitle }}</h3>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs font-mono">
              <div class="p-4 bg-card border border-subtle rounded-xl space-y-2">
                <div class="flex items-center gap-2 text-brand-periwinkle font-bold">
                  <Activity class="w-4 h-4" />
                  <span>1. {{ t.archStep1Title }}</span>
                </div>
                <p class="text-xs text-text-secondary font-sans leading-relaxed">
                  {{ t.archStep1Desc }}
                </p>
              </div>

              <div class="p-4 bg-card border border-subtle rounded-xl space-y-2">
                <div class="flex items-center gap-2 text-amber-400 font-bold">
                  <ShieldCheck class="w-4 h-4" />
                  <span>2. {{ t.archStep2Title }}</span>
                </div>
                <p class="text-xs text-text-secondary font-sans leading-relaxed">
                  {{ t.archStep2Desc }}
                </p>
              </div>

              <div class="p-4 bg-card border border-subtle rounded-xl space-y-2">
                <div class="flex items-center gap-2 text-emerald-400 font-bold">
                  <Send class="w-4 h-4" />
                  <span>3. {{ t.archStep3Title }}</span>
                </div>
                <p class="text-xs text-text-secondary font-sans leading-relaxed">
                  {{ t.archStep3Desc }}
                </p>
              </div>
            </div>

            <!-- Architecture Details Table -->
            <div class="overflow-x-auto pt-2">
              <table class="w-full text-left text-xs text-text-secondary">
                <thead class="bg-card font-mono text-[10px] uppercase text-text-secondary">
                  <tr>
                    <th class="py-2.5 px-3">{{ t.thComponent }}</th>
                    <th class="py-2.5 px-3">{{ t.thTech }}</th>
                    <th class="py-2.5 px-3">{{ t.thResponsibility }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-subtle font-sans text-xs">
                  <tr>
                    <td class="py-3 px-3 font-bold text-text-main font-mono">Frontend Web</td>
                    <td class="py-3 px-3 text-brand-periwinkle font-mono">Vue 3 + Vite + Tailwind</td>
                    <td class="py-3 px-3 text-text-secondary leading-relaxed">{{ t.archRow1 }}</td>
                  </tr>
                  <tr>
                    <td class="py-3 px-3 font-bold text-text-main font-mono">Backend Engine</td>
                    <td class="py-3 px-3 text-emerald-400 font-mono">Golang (Gin Framework)</td>
                    <td class="py-3 px-3 text-text-secondary leading-relaxed">{{ t.archRow2 }}</td>
                  </tr>
                  <tr>
                    <td class="py-3 px-3 font-bold text-text-main font-mono">Database</td>
                    <td class="py-3 px-3 text-sky-400 font-mono">PostgreSQL</td>
                    <td class="py-3 px-3 text-text-secondary leading-relaxed">{{ t.archRow3 }}</td>
                  </tr>
                  <tr>
                    <td class="py-3 px-3 font-bold text-text-main font-mono">WhatsApp Gateway</td>
                    <td class="py-3 px-3 text-status-up font-mono">Node.js (Baileys)</td>
                    <td class="py-3 px-3 text-text-secondary leading-relaxed">{{ t.archRow4 }}</td>
                  </tr>
                  <tr>
                    <td class="py-3 px-3 font-bold text-text-main font-mono">Queue &amp; Rate-Limiter</td>
                    <td class="py-3 px-3 text-red-400 font-mono">Redis (Asynq)</td>
                    <td class="py-3 px-3 text-text-secondary leading-relaxed">{{ t.archRow5 }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <!-- ══════════════════════════════════════════════════════════════════════
             SECTION 4: Troubleshooting & Diagnostics (Penanganan Masalah)
             ══════════════════════════════════════════════════════════════════════ -->
        <section v-else-if="activeSection === 'troubleshooting'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Wrench class="w-4 h-4 text-brand-periwinkle" />
                {{ t.troubleHeader }}
              </h2>
              <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.troubleSubheader }}</p>
            </div>
          </div>

          <div class="space-y-4">
            <div class="p-5 bg-surface border border-subtle rounded-2xl space-y-3 shadow-xl">
              <div class="flex items-center gap-2 font-mono font-bold text-amber-400 text-xs">
                <AlertTriangle class="w-4 h-4" />
                <span>1. {{ t.trouble1Title }}</span>
              </div>
              <p class="text-xs text-text-secondary leading-relaxed font-sans">
                <strong>{{ t.causeLabel }}:</strong> {{ t.trouble1Cause }}<br>
                <strong>{{ t.solutionLabel }}:</strong>
              </p>
              <ul class="list-disc list-inside space-y-1.5 pl-2 text-xs font-sans text-text-secondary leading-relaxed">
                <li>{{ t.trouble1Sol1 }}</li>
                <li>{{ t.trouble1Sol2 }}</li>
                <li>{{ t.trouble1Sol3 }}</li>
              </ul>
            </div>

            <div class="p-5 bg-surface border border-subtle rounded-2xl space-y-3 shadow-xl">
              <div class="flex items-center gap-2 font-mono font-bold text-red-400 text-xs">
                <AlertCircle class="w-4 h-4" />
                <span>2. {{ t.trouble2Title }}</span>
              </div>
              <p class="text-xs text-text-secondary leading-relaxed font-sans">
                <strong>{{ t.causeLabel }}:</strong> {{ t.trouble2Cause }}<br>
                <strong>{{ t.solutionLabel }}:</strong>
              </p>
              <ul class="list-disc list-inside space-y-1.5 pl-2 text-xs font-sans text-text-secondary leading-relaxed">
                <li>{{ t.trouble2Sol1 }}</li>
                <li>{{ t.trouble2Sol2 }}</li>
              </ul>
            </div>

            <div class="p-5 bg-surface border border-subtle rounded-2xl space-y-3 shadow-xl">
              <div class="flex items-center gap-2 font-mono font-bold text-brand-periwinkle text-xs">
                <Sliders class="w-4 h-4" />
                <span>3. {{ t.trouble3Title }}</span>
              </div>
              <p class="text-xs text-text-secondary leading-relaxed font-sans">
                <strong>{{ t.causeLabel }}:</strong> {{ t.trouble3Cause }}<br>
                <strong>{{ t.solutionLabel }}:</strong>
              </p>
              <ul class="list-disc list-inside space-y-1.5 pl-2 text-xs font-sans text-text-secondary leading-relaxed">
                <li>{{ t.trouble3Sol1 }}</li>
                <li>{{ t.trouble3Sol2 }}</li>
                <li>{{ t.trouble3Sol3 }}</li>
              </ul>
            </div>
          </div>
        </section>

        <!-- ══════════════════════════════════════════════════════════════════════
             SECTION 5: Support & Contact (Kontak Bantuan)
             ══════════════════════════════════════════════════════════════════════ -->
        <section v-else-if="activeSection === 'contact'" class="space-y-6 animate-fadeIn">
          <div class="flex items-center justify-between border-b border-subtle pb-3">
            <div>
              <h2 class="text-sm font-extrabold text-text-main font-mono flex items-center gap-2">
                <Headphones class="w-4 h-4 text-brand-periwinkle" />
                {{ t.contactHeader }}
              </h2>
              <p class="text-xs text-text-secondary mt-0.5 font-sans">{{ t.contactSubheader }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="p-6 bg-surface border border-subtle rounded-2xl space-y-3 shadow-xl">
              <div class="w-10 h-10 rounded-xl bg-brand-periwinkle/15 border border-brand-periwinkle/30 flex items-center justify-center text-brand-periwinkle">
                <Send class="w-5 h-5" />
              </div>
              <h3 class="text-sm font-bold text-text-main font-mono">{{ t.contact1Title }}</h3>
              <p class="text-xs font-mono text-brand-periwinkle select-all font-bold">noc.alerts@jabarprov.go.id</p>
              <p class="text-xs text-text-secondary leading-relaxed font-sans">
                {{ t.contact1Desc }}
              </p>
            </div>

            <div class="p-6 bg-surface border border-subtle rounded-2xl space-y-3 shadow-xl">
              <div class="w-10 h-10 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
                <Headphones class="w-5 h-5" />
              </div>
              <h3 class="text-sm font-bold text-text-main font-mono">{{ t.contact2Title }}</h3>
              <p class="text-xs font-mono text-emerald-400 font-bold">Ext: 4401 / 4402 (24/7)</p>
              <p class="text-xs text-text-secondary leading-relaxed font-sans">
                {{ t.contact2Desc }}
              </p>
            </div>
          </div>
        </section>

      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useLangStore, type Language } from '../stores/langStore';
import {
  BookOpen,
  HelpCircle,
  Network,
  Wrench,
  Headphones,
  Search,
  X,
  LayoutGrid,
  Server,
  AlertTriangle,
  AlertCircle,
  Sliders,
  Send,
  Activity,
  ShieldCheck,
  ChevronDown
} from 'lucide-vue-next';

type SectionKey = 'guides' | 'faq' | 'architecture' | 'troubleshooting' | 'contact';

const langStore = useLangStore();
const lang = computed<Language>(() => langStore.currentLang);

const activeSection = ref<SectionKey>('guides');
const searchQuery = ref('');
const selectedFaqTag = ref<string>('all');
const openedFaqs = ref<number[]>([0, 1]); // default open first two FAQs

const faqTags = [
  { key: 'notifikasi', label: { id: 'Notifikasi', en: 'Notifications' } },
  { key: 'akun', label: { id: 'Akun & Sesi', en: 'Account & Sessions' } },
  { key: 'insiden', label: { id: 'Insiden', en: 'Incidents' } },
  { key: 'keamanan', label: { id: 'Keamanan 2FA', en: 'Security & 2FA' } },
  { key: 'perangkat', label: { id: 'Perangkat', en: 'Devices' } },
  { key: 'laporan', label: { id: 'Laporan SLA', en: 'SLA Reports' } },
  { key: 'hak-akses', label: { id: 'Hak Akses RBAC', en: 'RBAC Access' } }
];

// UI Text Dictionary for Indonesian (ID) and English (EN)
const dict = {
  id: {
    badge: 'SANOC Knowledge Base & Bantuan',
    title: 'Pusat Bantuan & Panduan Pengguna',
    subtitle: 'Dokumentasi lengkap operasional sistem monitoring infrastruktur jaringan SANOC Jawa Barat, panduan alur kerja (User Guides), FAQ, arsitektur alert gateway, serta penanganan masalah (Troubleshooting).',
    searchPlaceholder: 'Cari panduan, topik, atau FAQ...',
    navHeader: 'Daftar Materi',
    quickHelpTitle: 'Butuh Bantuan Cepat?',
    quickHelpDesc: 'Hubungi Tim On-Call NOC Diskominfo Jabar 24/7 untuk eskalasi darurat.',
    allCategories: 'Semua Kategori',
    guidesHeader: 'PANDUAN OPERASIONAL PENGGUNA (USER GUIDES)',
    guidesSubheader: 'Langkah demi langkah penggunaan fitur utama sistem monitoring SANOC.',
    guide1Title: 'Membaca Dashboard & Live Feed Real-Time',
    guide1Sub: 'Pemantauan visual status node, availability SLA, dan flapping alert.',
    guide1Desc: 'Halaman Dashboard menyajikan ringkasan instan kondisi seluruh jaringan infrastruktur yang dipantau:',
    guide1Bullet1: 'Menampilkan total node aktif, jumlah node UP (hijau), jumlah node DOWN (merah), serta persentase Overall Uptime SLA.',
    guide1Bullet2: 'Memperbarui perubahan status perangkat (state transition) dan pengiriman notifikasi secara instan via WebSocket tanpa reload.',
    guide1Bullet3: 'Menampilkan daftar node yang sering berganti status dalam 7 hari terakhir sebagai indikasi gangguan intermiten atau kabel loss.',
    guide1Bullet4: 'Menjalankan siklus polling ICMP manual seketika ke seluruh perangkat tanpa menunggu jeda scheduler.',
    guide2Title: 'Manajemen Perangkat & Fitur Kelola Massal',
    guide2Sub: 'Penambahan node, import data batch, dan konfigurasi massal.',
    guide2Desc: 'Pada menu Devices, Anda dapat mengelola seluruh perangkat jaringan:',
    guide2Card1Title: 'Mode Tampilan',
    guide2Card1Desc: 'Gunakan toggle Grouped by Location untuk mengelompokkan perangkat per site/gedung, atau Flat List untuk daftar tabel menyeluruh.',
    guide2Card2Title: 'Kelola Massal (Bulk Edit)',
    guide2Card2Desc: 'Klik tombol Kelola Massal di bilah atas untuk memunculkan checkbox, pilih beberapa perangkat, lalu buka Drawer Kanan untuk ubah lokasi, tipe, atau hapus massal.',
    guide2Card3Title: 'Import CSV / Excel',
    guide2Card3Desc: 'Gunakan tombol Import CSV / Excel untuk mendaftarkan ratusan node sekaligus menggunakan file template spreadsheet yang telah disediakan.',
    guide2Card4Title: 'Detail & Riwayat Node',
    guide2Card4Desc: 'Klik baris perangkat untuk melihat grafik latency respons time, riwayat uptime 24 jam, perangkat lain di lokasi yang sama, dan riwayat insiden.',
    guide3Title: 'Penanganan Insiden & Audit Log Notifikasi',
    guide3Sub: 'Siklus hidup tiket insiden, konfirmasi kegagalan berturut-turut, dan auto-resolve.',
    guide3Desc: 'Sistem SANOC menggunakan algoritma Debounce & Flap Reuse untuk menjaga akurasi tiket gangguan:',
    guide3Bullet1: 'Insiden hanya dibuat jika perangkat gagal merespons ping sebanyak ambang batas konfirmasi berturut-turut (default: 3 kali).',
    guide3Bullet2: 'Saat node kembali online, tiket insiden otomatis diselesaikan (RESOLVED) dan durasi downtime dicatat.',
    guide3Bullet3: 'Jika node yang baru saja pulih kembali down dalam rentang waktu singkat (default: 10 menit), sistem menyambungkan timeline tiket sebelumnya.',
    guide3Bullet4: 'Pada halaman detail insiden, tabel Notification Audit Log mencatat rincian pengiriman pesan ke WhatsApp dan status fallback Telegram.',
    guide4Title: 'Konfigurasi Alert Gateway & Autentikasi 2FA',
    guide4Sub: 'Integrasi WhatsApp Baileys, Telegram Bot, rate limit, dan Two-Factor Auth.',
    guide4Desc: 'Pengaturan gateway dan keamanan dapat disesuaikan pada halaman Settings dan Profil Pengguna:',
    guide4Bullet1: 'Hubungkan akun WhatsApp NOC dengan menekan tombol QR Reconnect lalu scan kode QR di aplikasi WhatsApp smartphone Anda.',
    guide4Bullet2: 'Masukkan Bot Token dan Channel Chat ID. Jika WhatsApp terkendala, sistem otomatis mengalihkan alert ke Telegram.',
    guide4Bullet3: 'Buka menu Profile -> klik Enable 2FA -> scan QR code dengan Google Authenticator -> masukkan 6-digit kode verifikasi.',
    faqHeader: 'FREQUENTLY ASKED QUESTIONS (FAQ & QNA)',
    faqSubheader: 'Jawaban atas pertanyaan umum seputar pengoperasian dan fitur sistem SANOC.',
    faqCounterLabel: 'Tanya Jawab',
    categoryTag: 'Kategori',
    noResultTitle: 'Pertanyaan Tidak Ditemukan',
    noResultDesc: 'Coba gunakan kata kunci pencarian yang lain atau pilih kategori Semua.',
    archHeader: 'ARSITEKTUR & MEKANISME MONITORING',
    archSubheader: 'Struktur kerja backend poller, websocket realtime, dan notifikasi alert.',
    archFlowTitle: 'Topologi Alur Kerja Sistem SANOC',
    archStep1Title: 'ICMP Poller Engine',
    archStep1Desc: 'Worker pool Go melakukan probe ICMP ping non-blocking secara batch paralel ke seluruh IP node sesuai interval konfigurasi.',
    archStep2Title: 'State & Debounce',
    archStep2Desc: 'Status dicocokkan dengan threshold kegagalan. Jika down terkonfirmasi, sistem mencatat riwayat ke PostgreSQL dan memancarkan WebSocket.',
    archStep3Title: 'Asynq Alert Pipeline',
    archStep3Desc: 'Notifikasi dialirkan ke Redis queue dengan rate-limit spacing menuju WhatsApp Baileys Sidecar, dengan fallback otomatis ke Telegram Bot.',
    thComponent: 'Komponen',
    thTech: 'Teknologi',
    thResponsibility: 'Tanggung Jawab & Fungsi',
    archRow1: 'Dashboard visual, live websocket charts, kelola inventaris massal, audit logs.',
    archRow2: 'ICMP probe engine, REST API, RBAC permission matrix, WebSocket hub, report generator.',
    archRow3: 'Penyimpanan relasional perangkat, tiket insiden, riwayat status, konfigurasi sistem.',
    archRow4: 'Sidecar socket gateway untuk broadcast notifikasi instan ke grup/nomor operator.',
    archRow5: 'Antrean pesan asinkron, retry backoff, dan pembatasan frekuensi pengiriman pesan.',
    troubleHeader: 'PANDUAN TROUBLESHOOTING & PENANGANAN KENDALA',
    troubleSubheader: 'Solusi cepat saat menghadapi kendala operasional atau error status.',
    causeLabel: 'Penyebab',
    solutionLabel: 'Solusi Rekomendasi',
    trouble1Title: 'WhatsApp Gateway Berstatus Disconnected / Sering Putus',
    trouble1Cause: 'Sesi WhatsApp pada smartphone logout, aplikasi WhatsApp di smartphone ditutup oleh battery saver, atau sidecar Node.js belum aktif.',
    trouble1Sol1: 'Buka menu Settings -> pilih tab Gateways & Alerts.',
    trouble1Sol2: 'Klik tombol QR Reconnect lalu scan kode QR yang tampil menggunakan WhatsApp Linked Devices pada smartphone.',
    trouble1Sol3: 'Pastikan smartphone memiliki koneksi internet yang stabil dan izinkan WhatsApp berjalan di latar belakang.',
    trouble2Title: 'Perangkat Terindikasi DOWN Palsu Padahal Fisik Hidup',
    trouble2Cause: 'ICMP Ping diblokir oleh Windows Firewall / Access Control List router, atau jeda respons jaringan lebih lambat dari batas timeout.',
    trouble2Sol1: 'Buka firewall pada perangkat target dan aktifkan File and Printer Sharing (Echo Request - ICMPv4-In).',
    trouble2Sol2: 'Buka Settings -> Engine & Thresholds dan sesuaikan Consecutive ICMP Checks ke nilai yang lebih tinggi (3 atau 4 checks).',
    trouble3Title: 'MAC Address Perangkat Tidak Terdeteksi di Subnet Berbeda',
    trouble3Cause: 'Protokol ARP standar hanya bekerja pada Layer-2 (satu subnet lokal) dan tidak dapat melintasi router VLAN.',
    trouble3Sol1: 'Buka menu Settings -> pilih tab Core Switch & SNMP.',
    trouble3Sol2: 'Masukkan IP Core Router / Switch Gateway dan Community String (misal: public).',
    trouble3Sol3: 'Sistem akan men-query tabel ARP Layer-3 dari router melalui OID ipNetToMediaPhysAddress secara otomatis.',
    contactHeader: 'KONTAK LAYANAN & HELPDESK SANOC',
    contactSubheader: 'Informasi saluran dukungan teknis dan kontak darurat operasional.',
    contact1Title: 'Email Helpdesk NOC',
    contact1Desc: 'Kirimkan tiket kendala sistem, permohonan penambahan subnet IP baru, atau eskalasi masalah teknis.',
    contact2Title: 'Hotline NOC Jabar Regional',
    contact2Desc: 'Layanan telepon siaga darurat untuk insiden kritis (Major Outage) di lingkungan Gedung Sate dan OPD Jawa Barat.'
  },
  en: {
    badge: 'SANOC Knowledge Base & Support',
    title: 'Help Center & User Guides',
    subtitle: 'Comprehensive documentation for SANOC West Java network infrastructure monitoring system, workflow user guides, FAQs, alert gateway architecture, and troubleshooting procedures.',
    searchPlaceholder: 'Search guides, topics, or FAQs...',
    navHeader: 'Table of Contents',
    quickHelpTitle: 'Need Quick Support?',
    quickHelpDesc: 'Contact the 24/7 West Java Diskominfo NOC On-Call Team for emergency escalation.',
    allCategories: 'All Categories',
    guidesHeader: 'OPERATIONAL USER GUIDES',
    guidesSubheader: 'Step-by-step walkthroughs for mastering all SANOC monitoring features.',
    guide1Title: 'Reading Dashboard & Real-Time Live Feed',
    guide1Sub: 'Visual monitoring of node states, SLA uptime availability, and flapping alerts.',
    guide1Desc: 'The Dashboard presents an instant operational overview of all monitored network nodes:',
    guide1Bullet1: 'Displays total active nodes, UP count (green), DOWN/Outage count (red), and overall SLA percentage.',
    guide1Bullet2: 'Streams real-time state transitions and alert dispatches instantly via WebSocket without browser reloads.',
    guide1Bullet3: 'Highlights nodes with frequent state changes over the last 7 days to isolate intermittent physical link issues.',
    guide1Bullet4: 'Triggers an immediate manual ICMP probe cycle across all devices without waiting for the scheduler.',
    guide2Title: 'Devices Management & Bulk Operations',
    guide2Sub: 'Node provisioning, batch spreadsheet imports, and bulk configuration drawer.',
    guide2Desc: 'Manage your comprehensive infrastructure inventory under the Devices menu:',
    guide2Card1Title: 'View Modes',
    guide2Card1Desc: 'Use the Grouped by Location toggle to organize devices per facility/rack, or Flat List for a full searchable table.',
    guide2Card2Title: 'Bulk Operations (Bulk Edit)',
    guide2Card2Desc: 'Click the Bulk Operations toggle to reveal selection checkboxes, select multiple devices, and open the Right Drawer to change locations, types, or delete in bulk.',
    guide2Card3Title: 'Import CSV / Excel',
    guide2Card3Desc: 'Use the Import CSV / Excel modal to onboard hundreds of network nodes simultaneously using the provided template.',
    guide2Card4Title: 'Node Details & History',
    guide2Card4Desc: 'Click any device row to view its response latency charts, 24h uptime timeline, sibling devices at the same location, and past incident records.',
    guide3Title: 'Incident Management & Flap Reuse Lifecycle',
    guide3Sub: 'Incident lifecycle, consecutive failure thresholds, and automatic resolution.',
    guide3Desc: 'SANOC applies an intelligent Debounce & Flap Reuse algorithm to ensure alarm accuracy:',
    guide3Bullet1: 'An incident is created only after a node fails ping probes for a configured consecutive threshold (default: 3 times).',
    guide3Bullet2: 'When a node recovers, the incident ticket is automatically RESOLVED and the total outage duration is computed.',
    guide3Bullet3: 'If a recently recovered node goes DOWN again within a short time window (default: 10 minutes), the previous incident timeline is continued rather than duplicated.',
    guide3Bullet4: 'On the Incident Detail page, the Notification Audit Log displays WhatsApp delivery receipts and Telegram fallback statuses (including Skipped status).',
    guide4Title: 'Alert Gateways & 2FA Security Configuration',
    guide4Sub: 'WhatsApp Baileys, Telegram Bot fallback, rate limiting, and Two-Factor Authentication.',
    guide4Desc: 'Configure notification routing and account security in Settings and User Profile:',
    guide4Bullet1: 'Pair the NOC WhatsApp account by clicking QR Reconnect and scanning the QR code with your smartphone WhatsApp app.',
    guide4Bullet2: 'Specify Bot Token and Chat ID. When WhatsApp is unavailable, alert dispatches seamlessly failover to Telegram.',
    guide4Bullet3: 'Navigate to Profile -> click Enable 2FA -> scan the QR code with Google Authenticator -> submit the 6-digit OTP.',
    faqHeader: 'FREQUENTLY ASKED QUESTIONS (FAQ & QNA)',
    faqSubheader: 'Clear answers to common questions regarding SANOC operations and system behaviors.',
    faqCounterLabel: 'Q&As',
    categoryTag: 'Category',
    noResultTitle: 'No Matching Questions Found',
    noResultDesc: 'Try searching with different keywords or select All Categories.',
    archHeader: 'MONITORING ARCHITECTURE & MECHANISM',
    archSubheader: 'Internal mechanics of the Go poller, real-time WebSocket hub, and alert queues.',
    archFlowTitle: 'SANOC End-to-End System Workflow',
    archStep1Title: 'ICMP Poller Engine',
    archStep1Desc: 'Go worker pool executes non-blocking batch parallel ICMP ping probes across all node IPs based on configured intervals.',
    archStep2Title: 'State & Debounce Engine',
    archStep2Desc: 'Checks responses against failure thresholds. Confirmed outages persist to PostgreSQL and broadcast over WebSocket.',
    archStep3Title: 'Asynq Alert Pipeline',
    archStep3Desc: 'Alerts stream into a Redis queue with rate-limiting spacing to the Baileys WhatsApp sidecar, falling back to Telegram.',
    thComponent: 'Component',
    thTech: 'Technology',
    thResponsibility: 'Role & Responsibility',
    archRow1: 'Visual dashboard, live websocket charts, bulk inventory operations, audit logs.',
    archRow2: 'ICMP probe engine, REST API, RBAC permission matrix, WebSocket hub, report generator.',
    archRow3: 'Relational storage for devices, incidents, historical timeline, system configuration.',
    archRow4: 'Sidecar socket gateway for broadcasting instant outage alerts to operator groups.',
    archRow5: 'Asynchronous task queue, retry backoff, and transmission rate-limiting spacing.',
    troubleHeader: 'TROUBLESHOOTING & DIAGNOSTIC GUIDE',
    troubleSubheader: 'Actionable solutions for resolving common operational anomalies and errors.',
    causeLabel: 'Root Cause',
    solutionLabel: 'Recommended Resolution',
    trouble1Title: 'WhatsApp Gateway Status Disconnected or Frequently Dropping',
    trouble1Cause: 'Smartphone session logged out, smartphone power saver suspended WhatsApp, or Node.js sidecar service is inactive.',
    trouble1Sol1: 'Go to Settings -> Gateways & Alerts tab.',
    trouble1Sol2: 'Click QR Reconnect and scan the QR code using WhatsApp Linked Devices on your phone.',
    trouble1Sol3: 'Ensure the smartphone has steady internet connectivity and disable battery optimization for WhatsApp.',
    trouble2Title: 'Device Displays False DOWN Alarm While Physically Alive',
    trouble2Cause: 'ICMP echo requests blocked by Windows Firewall / router ACL, or network latency exceeds timeout threshold.',
    trouble2Sol1: 'Allow File and Printer Sharing (Echo Request - ICMPv4-In) in the target OS firewall rules.',
    trouble2Sol2: 'Open Settings -> Engine & Thresholds and increase Consecutive ICMP Checks (e.g. 3 or 4 checks).',
    trouble3Title: 'Device MAC Address Missing on Cross-Subnet / VLAN Nodes',
    trouble3Cause: 'Standard ARP resolution operates strictly on Layer-2 local broadcast domains and cannot traverse IP routers.',
    trouble3Sol1: 'Open Settings -> Core Switch & SNMP tab.',
    trouble3Sol2: 'Provide your Layer-3 Core Switch IP address and SNMP Community String (e.g., public).',
    trouble3Sol3: 'SANOC queries the router ARP routing table via OID ipNetToMediaPhysAddress automatically.',
    contactHeader: 'SANOC HELPDESK & SUPPORT CONTACTS',
    contactSubheader: 'Technical assistance channels and emergency operational contacts.',
    contact1Title: 'NOC Helpdesk Email',
    contact1Desc: 'Submit system anomaly tickets, request subnet expansion, or technical escalation.',
    contact2Title: 'Regional NOC Emergency Hotline',
    contact2Desc: '24/7 emergency dispatch line for major outages across West Java government regional infrastructure.'
  }
};

const t = computed(() => dict[lang.value]);

interface CategoryItem {
  id: SectionKey;
  label: Record<Language, string>;
  description: Record<Language, string>;
  icon: any;
}

const categories: CategoryItem[] = [
  {
    id: 'guides',
    label: { id: 'Panduan Pengguna', en: 'User Guides' },
    description: { id: 'Panduan operasional sistem', en: 'Operational system workflows' },
    icon: BookOpen
  },
  {
    id: 'faq',
    label: { id: 'FAQ & QnA', en: 'FAQ & QnA' },
    description: { id: 'Tanya jawab seputar fitur', en: 'Common questions & answers' },
    icon: HelpCircle
  },
  {
    id: 'architecture',
    label: { id: 'Arsitektur Sistem', en: 'System Architecture' },
    description: { id: 'Mekanisme & alur poller', en: 'Poller mechanics & flows' },
    icon: Network
  },
  {
    id: 'troubleshooting',
    label: { id: 'Troubleshooting', en: 'Troubleshooting' },
    description: { id: 'Diagnostik kendala teknis', en: 'Technical diagnostic guides' },
    icon: Wrench
  },
  {
    id: 'contact',
    label: { id: 'Kontak Bantuan', en: 'Support Contacts' },
    description: { id: 'Helpdesk & nomor siaga', en: 'Helpdesk & emergency hotlines' },
    icon: Headphones
  }
];

interface FAQItem {
  key: string;
  tagKey: string;
  question: Record<Language, string>;
  answer: Record<Language, string>;
  tag: Record<Language, string>;
}

const faqList: FAQItem[] = [
  {
    key: 'telegram-skipped',
    tagKey: 'notifikasi',
    question: {
      id: 'Mengapa status notifikasi Telegram berstatus "Skipped"?',
      en: 'Why is the Telegram notification status marked as "Skipped"?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Sistem SANOC menerapkan arsitektur notifikasi <strong>Primary &amp; Fallback</strong> yang cerdas:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>WhatsApp (Primer):</strong> Jalur utama pengiriman broadcast peringatan gangguan.</li>
          <li><strong>Telegram (Cadangan):</strong> Berfungsi sebagai failover otomatis jika transmisi WhatsApp gagal atau server Baileys terputus.</li>
          <li><strong>Status "Skipped":</strong> Jika WhatsApp berhasil mengirimkan pesan 100%, sistem sengaja melewati pengiriman ke Telegram agar tidak terjadi duplikasi alarm (spam) bagi operator NOC.</li>
        </ul>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">SANOC employs an intelligent <strong>Primary &amp; Fallback</strong> alert architecture:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>WhatsApp (Primary):</strong> The primary dispatch channel for outage alerts.</li>
          <li><strong>Telegram (Fallback):</strong> Operates as an automated contingency channel if WhatsApp fails or disconnects.</li>
          <li><strong>"Skipped" Status:</strong> When WhatsApp dispatches successfully, Telegram is deliberately skipped to prevent duplicate alarm noise.</li>
        </ul>
      `
    },
    tag: { id: 'Notifikasi', en: 'Notifications' }
  },
  {
    key: 'session-duration',
    tagKey: 'akun',
    question: {
      id: 'Berapa lama sesi login akun bertahan sebelum harus login ulang?',
      en: 'How long does a user login session remain valid before expiration?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Sesi login akun SANOC telah ditingkatkan menjadi <strong>24 Jam Penuh</strong>:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>JWT Token Lifetime:</strong> Berlaku selama 86.400 detik (24 jam) sejak waktu login berhasil.</li>
          <li><strong>HttpOnly Cookie:</strong> Menggunakan perlindungan cookie HttpOnly yang aman dari serangan XSS.</li>
          <li><strong>Operator Continuity:</strong> Memastikan staf NOC tidak ter-logout secara mendadak di tengah shift pemantauan operasional.</li>
        </ul>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">SANOC authentication sessions are configured for <strong>24 Full Hours</strong>:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>JWT Token Lifetime:</strong> Valid for 86,400 seconds (24 hours) from successful login.</li>
          <li><strong>HttpOnly Cookie:</strong> Protected against XSS exploitation via secure HttpOnly cookie encapsulation.</li>
          <li><strong>Shift Continuity:</strong> Prevents sudden session expiries during continuous operational monitoring shifts.</li>
        </ul>
      `
    },
    tag: { id: 'Akun & Sesi', en: 'Account & Sessions' }
  },
  {
    key: 'flap-reuse',
    tagKey: 'insiden',
    question: {
      id: 'Apa itu Flap Detection Reuse Window dan bagaimana cara kerjanya?',
      en: 'What is the Flap Detection Reuse Window and how does it work?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Mekanisme anti-flapping mencegah timbulnya ratusan tiket palsu saat koneksi perangkat tidak stabil:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>Kondisi Flapping:</strong> Terjadi ketika kabel longgar atau daya intermiten menyebabkan perangkat UP &rarr; DOWN &rarr; UP berulang kali.</li>
          <li><strong>Reuse Window (Default 10 Menit):</strong> Jika perangkat yang baru saja pulih kembali DOWN dalam kurun waktu 10 menit, sistem tidak membuat tiket baru melainkan <strong>melanjutkan tiket insiden yang sama</strong>.</li>
          <li><strong>Audit Timeline Utuh:</strong> Seluruh kronologis gangguan terekam dalam satu riwayat terpadu pada halaman detail insiden.</li>
        </ul>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">Anti-flapping logic prevents ticket flooding caused by unstable intermittent links:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>Flapping Condition:</strong> Occurs when loose physical cabling or power oscillations toggle nodes UP/DOWN repeatedly.</li>
          <li><strong>Reuse Window (Default 10 Min):</strong> If a recently resolved node drops DOWN again within 10 minutes, SANOC reopens and appends to the original incident ticket.</li>
          <li><strong>Unified Timeline:</strong> Preserves full chronological troubleshooting context in a single incident record.</li>
        </ul>
      `
    },
    tag: { id: 'Insiden', en: 'Incidents' }
  },
  {
    key: 'two-factor-auth',
    tagKey: 'keamanan',
    question: {
      id: 'Bagaimana cara mengaktifkan Two-Factor Authentication (2FA) dengan QR Code?',
      en: 'How do I activate Two-Factor Authentication (2FA) using QR Code?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Langkah aktivasi keamanan 2FA pada akun Anda:</p>
        <ol class="list-decimal list-inside space-y-1 text-text-secondary">
          <li>Klik foto profil/nama akun Anda di pojok kiri bawah untuk membuka halaman <strong>Profil Pengguna</strong>.</li>
          <li>Pada bagian <em>Two-Factor Authentication</em>, klik tombol <strong>Enable 2FA</strong>.</li>
          <li>Buka aplikasi autentikator (Google Authenticator, Microsoft Authenticator, atau Authy) di smartphone Anda.</li>
          <li>Pindai (scan) kode QR yang muncul di layar, lalu masukkan 6 digit angka OTP untuk mengonfirmasi aktivasi.</li>
        </ol>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">Follow these steps to activate 2FA on your account:</p>
        <ol class="list-decimal list-inside space-y-1 text-text-secondary">
          <li>Click your profile name/avatar in the bottom-left corner to open <strong>User Profile</strong>.</li>
          <li>Under the <em>Two-Factor Authentication</em> section, click <strong>Enable 2FA</strong>.</li>
          <li>Open your mobile authenticator app (Google Authenticator, Microsoft Authenticator, or Authy).</li>
          <li>Scan the visual QR code and enter the 6-digit confirmation OTP to complete activation.</li>
        </ol>
      `
    },
    tag: { id: 'Keamanan 2FA', en: 'Security & 2FA' }
  },
  {
    key: 'bulk-operations',
    tagKey: 'perangkat',
    question: {
      id: 'Bagaimana cara menggunakan fitur Kelola Massal (Bulk Edit) pada perangkat?',
      en: 'How do I utilize the Bulk Operations (Bulk Edit) tool for devices?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Fitur Kelola Massal mempermudah pengelolaan ratusan node sekaligus:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>Aktivasi Mode:</strong> Pada menu <strong>Devices</strong>, klik tombol <strong>Kelola Massal (Bulk)</strong> di bilah atas.</li>
          <li><strong>Pilih Perangkat:</strong> Centang kotak checkbox di sebelah kiri perangkat yang ingin dikonfigurasi.</li>
          <li><strong>Slide-Over Right Drawer:</strong> Klik tombol melayang <em>Buka Panel Konfigurasi Massal</em> di bagian bawah.</li>
          <li><strong>Aksi Massal:</strong> Pilih aksi yang diinginkan: Pindah Lokasi, Ubah Kategori Tipe, Force ICMP Poll Seketika, atau Hapus Massal.</li>
        </ul>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">Bulk Operations enables efficient management of multiple nodes simultaneously:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>Activate Mode:</strong> On the <strong>Devices</strong> page, toggle the <strong>Bulk Operations</strong> button in the top toolbar.</li>
          <li><strong>Select Devices:</strong> Select checkboxes beside target devices.</li>
          <li><strong>Slide-Over Drawer:</strong> Click the floating action bar button at the bottom to open the side configuration panel.</li>
          <li><strong>Execute Actions:</strong> Perform batch location moves, category updates, immediate ICMP force polling, or bulk deletions.</li>
        </ul>
      `
    },
    tag: { id: 'Perangkat', en: 'Devices' }
  },
  {
    key: 'export-reports',
    tagKey: 'laporan',
    question: {
      id: 'Bagaimana cara mengekspor laporan downtime dan ketersediaan SLA?',
      en: 'How do I generate and export SLA uptime and downtime reports?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Penyusunan laporan ketersediaan jaringan SLA:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li>Buka menu <strong>Reports</strong> pada sidebar navigasi kiri.</li>
          <li>Tentukan rentang tanggal pelaporan (<em>Harian, Mingguan, Bulanan, atau Custom Date Range</em>).</li>
          <li>Tinjau diagram distribusi downtime, persentase MTTR (Mean Time To Recovery), dan ketersediaan SLA per lokasi.</li>
          <li>Klik tombol <strong>Export PDF Report</strong> untuk format cetak resmi, atau <strong>Export Excel (XLSX)</strong> untuk olah data spreadsheet.</li>
        </ul>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">Generating formal SLA availability audit reports:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li>Navigate to the <strong>Reports</strong> tab in the sidebar.</li>
          <li>Select the audit date range (<em>Daily, Weekly, Monthly, or Custom Range</em>).</li>
          <li>Inspect MTTR metrics, downtime distribution charts, and per-location SLA availability.</li>
          <li>Click <strong>Export PDF Report</strong> for formal print layouts or <strong>Export Excel (XLSX)</strong> for raw spreadsheet analysis.</li>
        </ul>
      `
    },
    tag: { id: 'Laporan SLA', en: 'SLA Reports' }
  },
  {
    key: 'whatsapp-reconnect',
    tagKey: 'notifikasi',
    question: {
      id: 'Bagaimana cara menghubungkan kembali WhatsApp Gateway jika terputus?',
      en: 'How do I reconnect the WhatsApp Gateway when disconnected?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Prosedur penautan ulang sesi WhatsApp Baileys:</p>
        <ol class="list-decimal list-inside space-y-1 text-text-secondary">
          <li>Buka menu <strong>Settings</strong> &rarr; pilih tab <strong>Gateways &amp; Alerts</strong>.</li>
          <li>Klik tombol <strong>QR Reconnect</strong> pada kartu WhatsApp Gateway.</li>
          <li>Buka aplikasi WhatsApp di smartphone operator, masuk ke menu <strong>Perangkat Tertaut (Linked Devices)</strong>.</li>
          <li>Arahkan kamera smartphone ke kode QR yang muncul pada layar hingga status berubah menjadi <em>Connected (Hijau)</em>.</li>
        </ol>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">Re-authenticating the Baileys WhatsApp Gateway session:</p>
        <ol class="list-decimal list-inside space-y-1 text-text-secondary">
          <li>Open <strong>Settings</strong> &rarr; select the <strong>Gateways &amp; Alerts</strong> tab.</li>
          <li>Click <strong>QR Reconnect</strong> on the WhatsApp Gateway card.</li>
          <li>Open WhatsApp on your phone and go to <strong>Linked Devices</strong>.</li>
          <li>Scan the visual QR code until the status indicator turns green (<em>Connected</em>).</li>
        </ol>
      `
    },
    tag: { id: 'Notifikasi', en: 'Notifications' }
  },
  {
    key: 'rbac-permissions',
    tagKey: 'hak-akses',
    question: {
      id: 'Mengapa tab konfigurasi di menu Settings tidak muncul untuk akun saya?',
      en: 'Why are specific Settings configuration tabs hidden on my account?'
    },
    answer: {
      id: `
        <p class="font-medium text-text-main mb-1.5">Menu Settings dikontrol ketat oleh <strong>Role-Based Access Control (RBAC)</strong>:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>Akses Granular:</strong> Setiap kategori Settings (Notifikasi, Polling, Jaringan, Retensi, Lokasi, Manajemen User, dan Audit Log) memiliki izin independen.</li>
          <li><strong>Perlindungan 403:</strong> Jika role akun Anda (misal: <em>Anggota</em> atau <em>Pimpinan</em>) belum diizinkan oleh Super Admin, tab tersebut disembunyikan sepenuhnya dari antarmuka.</li>
          <li><strong>Eskalasi:</strong> Hubungi Administrator untuk memperbarui izin akun Anda melalui matriks <em>Access Control Matrix</em>.</li>
        </ul>
      `,
      en: `
        <p class="font-medium text-text-main mb-1.5">The Settings interface enforces strict <strong>Role-Based Access Control (RBAC)</strong>:</p>
        <ul class="list-disc list-inside space-y-1 text-text-secondary">
          <li><strong>Granular Permissions:</strong> Each category (Notifications, Polling, Network, Retention, Locations, Users, Audit) operates with an independent permission key.</li>
          <li><strong>403 Access Guard:</strong> If your role (e.g. <em>Staff</em> or <em>Executive</em>) lacks permission, the tab is securely hidden.</li>
          <li><strong>Request Access:</strong> Contact your Super Admin to update your role permissions in the <em>Access Control Matrix</em>.</li>
        </ul>
      `
    },
    tag: { id: 'Hak Akses RBAC', en: 'RBAC Access' }
  }
];

const filteredFaqs = computed(() => {
  let list = faqList;
  if (selectedFaqTag.value !== 'all') {
    list = list.filter((f) => f.tagKey === selectedFaqTag.value);
  }
  if (!searchQuery.value.trim()) return list;

  const q = searchQuery.value.toLowerCase();
  const currentLanguage = lang.value;
  return list.filter(
    (f) =>
      f.question[currentLanguage].toLowerCase().includes(q) ||
      f.answer[currentLanguage].toLowerCase().includes(q) ||
      f.tag[currentLanguage].toLowerCase().includes(q)
  );
});

function toggleFaq(idx: number) {
  const i = openedFaqs.value.indexOf(idx);
  if (i > -1) {
    openedFaqs.value.splice(i, 1);
  } else {
    openedFaqs.value.push(idx);
  }
}
</script>

<style scoped>
.faq-content :deep(strong) {
  color: var(--text-primary);
  font-weight: 600;
}
.faq-content :deep(code) {
  background-color: var(--bg-hover);
  color: var(--status-warning);
  padding: 0.15rem 0.35rem;
  border-radius: 0.375rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
}
.faq-content :deep(ul), .faq-content :deep(ol) {
  margin-top: 0.35rem;
}
.faq-content :deep(li) {
  margin-bottom: 0.25rem;
}
</style>
