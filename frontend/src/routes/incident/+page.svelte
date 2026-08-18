<script lang="ts">
  import { onMount } from 'svelte';
  import { fade, fly, slide } from 'svelte/transition';

  // API base URL
  const API_URL = 'http://localhost:8081';

  interface Incident {
    id?: number;
    date: string; // YYYY-MM-DD
    title: string;
    description: string;
    root_cause: string;
    ticket_count: number;
    updated_at?: string;
  }

  interface DetectedSpike {
    date: string;
    ticket_count: number;
  }

  // State Management
  let isLoading = $state(true);
  let savedIncidents = $state<Incident[]>([]);
  let detectedSpikes = $state<DetectedSpike[]>([]);
  
  // Form State
  let showEditModal = $state(false);
  let isEditingExisting = $state(false);
  let formIncident = $state<Incident>({
    date: '',
    title: '',
    description: '',
    root_cause: '',
    ticket_count: 0
  });

  let saveError = $state('');
  let isSaving = $state(false);

  // Derived: Detected spike days that haven't been archived yet
  const unarchivedSpikes = $derived(
    detectedSpikes.filter(spike => 
      !savedIncidents.some(saved => saved.date === spike.date)
    )
  );

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    isLoading = true;
    try {
      const [incidentsRes, spikesRes] = await Promise.all([
        fetch(`${API_URL}/incidents`),
        fetch(`${API_URL}/detected-incidents`)
      ]);

      if (incidentsRes.ok) {
        savedIncidents = await incidentsRes.json() || [];
      }
      if (spikesRes.ok) {
        detectedSpikes = await spikesRes.json() || [];
      }
    } catch (e) {
      console.error('Failed to fetch incident data:', e);
    } finally {
      isLoading = false;
    }
  }

  function startArchive(spike: DetectedSpike) {
    isEditingExisting = false;
    saveError = '';
    formIncident = {
      date: spike.date,
      title: `Incident: Spike on ${formatDateLabel(spike.date)}`,
      description: '',
      root_cause: '',
      ticket_count: spike.ticket_count
    };
    showEditModal = true;
  }

  function startEdit(incident: Incident) {
    isEditingExisting = true;
    saveError = '';
    formIncident = { ...incident };
    showEditModal = true;
  }

  function startCreateManual() {
    isEditingExisting = false;
    saveError = '';
    const today = new Date().toISOString().split('T')[0];
    formIncident = {
      date: today,
      title: '',
      description: '',
      root_cause: '',
      ticket_count: 0
    };
    showEditModal = true;
  }

  async function saveIncident() {
    if (!formIncident.date || !formIncident.title.trim()) {
      saveError = 'Date and Title are required.';
      return;
    }

    isSaving = true;
    saveError = '';
    try {
      const res = await fetch(`${API_URL}/incidents`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formIncident)
      });

      if (res.ok) {
        showEditModal = false;
        await loadData();
      } else {
        const errData = await res.json();
        saveError = errData.error || 'Failed to save incident.';
      }
    } catch (e) {
      saveError = 'Failed to connect to backend server.';
    } finally {
      isSaving = false;
    }
  }

  // Formatting helper
  function formatDateLabel(dateStr: string) {
    if (!dateStr) return '';
    const parts = dateStr.split('-');
    if (parts.length !== 3) return dateStr;
    const date = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function formatDateTime(dateTimeStr?: string) {
    if (!dateTimeStr) return '-';
    const date = new Date(dateTimeStr);
    if (isNaN(date.getTime())) return dateTimeStr;
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
</script>

<svelte:head>
  <title>IAM Incident Archive</title>
  <meta name="description" content="Manage, document, and review historical IAM ticket incidents and root causes." />
</svelte:head>

<!-- Background glow components -->
<div class="fixed inset-0 -z-50 bg-slate-950 overflow-hidden">
  <div class="absolute -top-40 -left-40 w-96 h-96 bg-rose-600/5 rounded-full blur-[128px]"></div>
  <div class="absolute top-1/3 -right-40 w-[500px] h-[500px] bg-indigo-650/10 rounded-full blur-[160px]"></div>
  <div class="absolute -bottom-40 left-1/3 w-[600px] h-[600px] bg-slate-900/10 rounded-full blur-[180px]"></div>
  <!-- Grid -->
  <div class="absolute inset-0 bg-[linear-gradient(to_right,#0f172a_1px,transparent_1px),linear-gradient(to_bottom,#0f172a_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] opacity-30"></div>
</div>

<div class="min-h-screen flex flex-col">
  
  <!-- Header -->
  <header class="border-b border-slate-900/60 bg-slate-950/80 backdrop-blur-xl sticky top-0 z-30 px-6 py-4 flex items-center justify-between">
    <div class="flex items-center gap-3">
      <!-- Logo -->
      <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-rose-600 to-indigo-500 flex items-center justify-center text-white shadow-md shadow-rose-600/20">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
      </div>
      <div>
        <h1 class="text-base font-bold text-white tracking-tight">IAM Incident Archive</h1>
        <p class="text-[10px] text-slate-500 font-mono leading-none">V1.0.0 // Incidents</p>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <nav class="hidden sm:flex items-center gap-1 bg-slate-900 border border-slate-850 p-1 rounded-xl text-xs font-semibold">
      <a href="/" class="px-3.5 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
        </svg>
        <span>Home</span>
      </a>
      <a href="/analysis" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Analytics</a>
      <a href="/leaderboard" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Leaderboard</a>
      <a href="/incident" class="px-4 py-2 rounded-lg bg-rose-600 text-white shadow-md shadow-rose-600/10 transition">Incidents</a>
    </nav>

    <div class="flex items-center gap-3">
      <button 
        onclick={startCreateManual}
        class="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 hover:border-slate-700/80 hover:bg-slate-850 text-slate-200 text-xs font-semibold focus:outline-none transition shadow"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4 text-rose-500">
          <path d="M10.75 4.75a.75.75 0 00-1.5 0v4.5h-4.5a.75.75 0 000 1.5h4.5v4.5a.75.75 0 001.5 0v-4.5h4.5a.75.75 0 000-1.5h-4.5v-4.5z" />
        </svg>
        <span>Create Incident</span>
      </button>
    </div>
  </header>

  <main class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">
    
    <div class="space-y-1">
      <h2 class="text-xl font-bold tracking-tight text-white">Historical Incident Archive</h2>
      <p class="text-xs text-slate-400 font-light">Document, analyze, and save root-cause logs for historical high-volume ticket spikes.</p>
    </div>

    <!-- Quick Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-4 flex flex-col justify-between shadow-lg relative overflow-hidden group">
        <div class="absolute -right-16 -top-16 w-32 h-32 bg-rose-500/5 rounded-full blur-xl"></div>
        <span class="text-[9px] font-bold text-slate-500 uppercase tracking-wider block">Archived Incidents</span>
        <span class="text-2xl font-black text-white tracking-tight mt-2 font-mono">{savedIncidents.length}</span>
      </div>

      <div class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-4 flex flex-col justify-between shadow-lg relative overflow-hidden group">
        <div class="absolute -right-16 -top-16 w-32 h-32 bg-amber-500/5 rounded-full blur-xl"></div>
        <span class="text-[9px] font-bold text-slate-500 uppercase tracking-wider block">Undocumented Spikes</span>
        <span class="text-2xl font-black text-amber-400 tracking-tight mt-2 font-mono">{unarchivedSpikes.length}</span>
      </div>

      <div class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-4 flex flex-col justify-between shadow-lg relative overflow-hidden group">
        <div class="absolute -right-16 -top-16 w-32 h-32 bg-indigo-500/5 rounded-full blur-xl"></div>
        <span class="text-[9px] font-bold text-slate-500 uppercase tracking-wider block">Spike Threshold</span>
        <span class="text-2xl font-black text-indigo-400 tracking-tight mt-2 font-mono">≥ 50 tickets/day</span>
      </div>
    </div>

    {#if isLoading}
      <!-- Loading State -->
      <div class="flex flex-col items-center justify-center py-32 space-y-4">
        <div class="relative w-16 h-16">
          <div class="absolute inset-0 border-4 border-rose-500/20 rounded-full"></div>
          <div class="absolute inset-0 border-4 border-t-rose-500 rounded-full animate-spin"></div>
        </div>
        <p class="text-sm text-slate-400 font-mono">Loading incident logbooks...</p>
      </div>

    {:else}
      <!-- Two Column Layout -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
        
        <!-- Left Side: Spikes needing documentation (Grid Span 4) -->
        <div class="lg:col-span-4 space-y-4">
          <div class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-5 shadow-xl space-y-4">
            <div>
              <h3 class="text-sm font-bold text-slate-200">Undocumented Volumetric Spikes</h3>
              <p class="text-xs text-slate-500 mt-0.5">Days with 50+ tickets that have not yet been archived.</p>
            </div>

            <div class="space-y-2.5 max-h-[500px] overflow-y-auto pr-1 scrollbar-thin scrollbar-thumb-slate-900">
              {#if unarchivedSpikes.length === 0}
                <div class="text-center text-slate-500 font-mono text-xs py-12 bg-slate-900/10 rounded-xl border border-dashed border-slate-900">
                  🎉 All spike days documented!
                </div>
              {:else}
                {#each unarchivedSpikes as spike}
                  <div class="p-3.5 rounded-xl bg-slate-900/30 border border-slate-900 hover:border-slate-800 hover:bg-slate-900/60 transition flex items-center justify-between gap-3">
                    <div class="space-y-1">
                      <span class="text-xs font-mono text-slate-350 font-bold block">{formatDateLabel(spike.date)}</span>
                      <span class="text-[10px] font-mono text-slate-500">{spike.ticket_count} tickets detected</span>
                    </div>
                    <button 
                      onclick={() => startArchive(spike)}
                      class="px-3 py-1.5 bg-rose-500/10 border border-rose-500/30 hover:bg-rose-500/20 text-rose-450 hover:text-rose-350 rounded-lg text-[10px] font-bold tracking-wide transition shrink-0 cursor-pointer"
                    >
                      Archive Day
                    </button>
                  </div>
                {/each}
              {/if}
            </div>
          </div>
        </div>

        <!-- Right Side: Saved Incident Archives (Grid Span 8) -->
        <div class="lg:col-span-8 space-y-4">
          <div class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-5 shadow-xl space-y-4">
            <div>
              <h3 class="text-sm font-bold text-slate-200">Incident Logs</h3>
              <p class="text-xs text-slate-550 mt-0.5">Historical catalog of logged outages, errors, and authentication degradations.</p>
            </div>

            <div class="space-y-3.5">
              {#if savedIncidents.length === 0}
                <div class="text-center text-slate-500 font-mono text-xs py-24 bg-slate-900/10 rounded-2xl border border-dashed border-slate-900">
                  No incident records found. Click "Create Incident" or select an undocumented spike to start.
                </div>
              {:else}
                {#each savedIncidents as incident}
                  <div class="p-5 rounded-2xl border border-slate-900 bg-slate-950/50 hover:border-slate-850 hover:bg-slate-900/10 transition-all duration-300 relative group flex flex-col justify-between gap-4">
                    <!-- Header -->
                    <div class="flex items-start justify-between gap-4">
                      <div class="space-y-1">
                        <div class="flex items-center gap-2">
                          <span class="px-2 py-0.5 rounded bg-rose-500/15 border border-rose-500/25 text-[10px] font-mono font-bold text-rose-400">
                            {formatDateLabel(incident.date)}
                          </span>
                          <span class="text-[10px] font-mono text-slate-500">
                            Volume: {incident.ticket_count} tickets
                          </span>
                        </div>
                        <h4 class="text-sm font-bold text-slate-200 group-hover:text-white transition">
                          {incident.title}
                        </h4>
                      </div>
                      
                      <button 
                        onclick={() => startEdit(incident)}
                        class="px-3 py-1.5 bg-slate-900 border border-slate-800 hover:border-slate-700/80 hover:bg-slate-850 text-slate-350 text-[10px] font-bold rounded-lg transition cursor-pointer shrink-0"
                      >
                        Edit Incident
                      </button>
                    </div>

                    <!-- Details -->
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs font-light leading-relaxed">
                      {#if incident.description}
                        <div class="space-y-1 bg-slate-950/40 p-3 rounded-xl border border-slate-900">
                          <span class="text-[9px] font-mono font-bold text-slate-500 uppercase tracking-widest block">Description / Impact</span>
                          <p class="text-slate-300 whitespace-pre-line">{incident.description}</p>
                        </div>
                      {/if}

                      {#if incident.root_cause}
                        <div class="space-y-1 bg-slate-950/40 p-3 rounded-xl border border-slate-900">
                          <span class="text-[9px] font-mono font-bold text-slate-500 uppercase tracking-widest block">Root Cause Analysis</span>
                          <p class="text-slate-300 whitespace-pre-line">{incident.root_cause}</p>
                        </div>
                      {/if}
                    </div>

                    <!-- Footer timestamp -->
                    <div class="text-[9px] font-mono text-slate-650 border-t border-slate-900/60 pt-2 flex justify-between items-center">
                      <span>Updated: {formatDateTime(incident.updated_at)}</span>
                      <span>DB Archive ID: #{incident.id}</span>
                    </div>
                  </div>
                {/each}
              {/if}
            </div>
          </div>
        </div>

      </div>
    {/if}

  </main>

  <!-- Footer -->
  <footer class="border-t border-slate-900/40 bg-slate-950 py-6 px-6 text-center text-[10px] text-slate-600 font-mono">
    IAM Ticket Dashboard // SQLite 3 & Gin API // Developed Pair-wise
  </footer>

</div>

<!-- Edit Incident Modal -->
{#if showEditModal}
  <!-- Overlay -->
  <div 
    onclick={() => showEditModal = false}
    class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-45"
    transition:fade
    role="button"
    tabindex="0"
  ></div>

  <!-- Modal Dialog -->
  <div 
    class="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-lg bg-slate-950 border border-slate-900 rounded-3xl p-6 shadow-2xl z-50 flex flex-col gap-5 overflow-hidden"
    transition:fly={{ y: 20, duration: 250 }}
  >
    <div class="flex items-center justify-between border-b border-slate-900/60 pb-3">
      <div>
        <h3 class="text-sm font-bold text-slate-200">
          {isEditingExisting ? 'Edit Incident Record' : 'Archive New Incident'}
        </h3>
        <p class="text-[10px] text-slate-500 font-mono mt-0.5">DATE: {formatDateLabel(formIncident.date)}</p>
      </div>
      <button 
        onclick={() => showEditModal = false}
        class="w-7 h-7 rounded-lg bg-slate-900 hover:bg-slate-800 text-slate-400 flex items-center justify-center border border-slate-800 cursor-pointer"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
          <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
        </svg>
      </button>
    </div>

    {#if saveError}
      <div class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs font-medium">
        {saveError}
      </div>
    {/if}

    <div class="space-y-4 text-xs">
      
      <!-- Date Picker (Only editable when manually creating and not archiving a detected date) -->
      {#if !isEditingExisting && !detectedSpikes.some(s => s.date === formIncident.date)}
        <div class="space-y-1.5">
          <label for="date-picker" class="font-semibold text-slate-350 block">Incident Date</label>
          <input 
            id="date-picker"
            type="date"
            bind:value={formIncident.date}
            class="w-full bg-slate-900 border border-slate-800 text-slate-200 rounded-xl px-3 py-2 text-xs focus:outline-none focus:border-rose-500/80"
          />
        </div>
      {:else}
        <input type="hidden" bind:value={formIncident.date} />
      {/if}

      <!-- Title Input -->
      <div class="space-y-1.5">
        <label for="title-input" class="font-semibold text-slate-350 block">Incident Title</label>
        <input 
          id="title-input"
          type="text" 
          bind:value={formIncident.title}
          placeholder="e.g. Onelogin SSO Server Failure" 
          class="w-full bg-slate-900 border border-slate-800 placeholder-slate-600 text-slate-200 rounded-xl px-3 py-2.5 text-xs focus:outline-none focus:border-rose-500/80"
        />
      </div>

      <!-- Ticket Count Input -->
      <div class="space-y-1.5">
        <label for="count-input" class="font-semibold text-slate-350 block">Impacted Ticket Volume</label>
        <input 
          id="count-input"
          type="number" 
          bind:value={formIncident.ticket_count}
          class="w-full bg-slate-900 border border-slate-800 text-slate-200 rounded-xl px-3 py-2 text-xs focus:outline-none focus:border-rose-500/80"
        />
      </div>

      <!-- Description Input -->
      <div class="space-y-1.5">
        <label for="desc-input" class="font-semibold text-slate-350 block">Description & Operational Impact</label>
        <textarea 
          id="desc-input"
          bind:value={formIncident.description}
          rows="3"
          placeholder="Explain what occurred and how it impacted operations..." 
          class="w-full bg-slate-900 border border-slate-800 placeholder-slate-600 text-slate-200 rounded-xl px-3 py-2 text-xs focus:outline-none focus:border-rose-500/80 resize-none"
        ></textarea>
      </div>

      <!-- Root Cause Input -->
      <div class="space-y-1.5">
        <label for="rc-input" class="font-semibold text-slate-350 block">Root Cause Analysis (RCA)</label>
        <textarea 
          id="rc-input"
          bind:value={formIncident.root_cause}
          rows="3"
          placeholder="Identify the underlying root cause of this incident..." 
          class="w-full bg-slate-900 border border-slate-800 placeholder-slate-600 text-slate-200 rounded-xl px-3 py-2 text-xs focus:outline-none focus:border-rose-500/80 resize-none"
        ></textarea>
      </div>

    </div>

    <!-- Actions -->
    <div class="flex items-center justify-end gap-3 pt-3 border-t border-slate-900/60 mt-2">
      <button 
        onclick={() => showEditModal = false}
        class="px-4 py-2 rounded-xl border border-slate-800 bg-slate-900/30 hover:bg-slate-900 text-slate-400 text-xs transition cursor-pointer"
      >
        Cancel
      </button>
      <button 
        onclick={saveIncident}
        disabled={isSaving}
        class="px-5 py-2 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold transition flex items-center gap-2 cursor-pointer shadow-lg shadow-rose-600/10"
      >
        {#if isSaving}
          <div class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></div>
        {/if}
        <span>{isEditingExisting ? 'Save Changes' : 'Archive Incident'}</span>
      </button>
    </div>
  </div>
{/if}
