<script lang="ts">
  import { onMount } from 'svelte';
  import { fade, fly, slide } from 'svelte/transition';

  // API base URL
  const API_URL = 'http://localhost:8081';

  interface Ticket {
    ticket_id: string;
    ticket_type: string;
    subject: string;
    priority: string;
    ticket_status: string;
    Assignee?: string;
    CreatedDate: string;
    ResolvedAt?: string;
    ClosedDate?: string;
    CompleteTime?: string;
    DetailDescription?: string;
  }

  // Reactive state using Runes
  let isLoading = $state(true);
  let tickets = $state<Ticket[]>([]);
  let months = $state<string[]>([]);
  let selectedMonth = $state<string>('');
  let isMonthDropdownOpen = $state(false);

  let isInitialized = $state(false);

  onMount(async () => {
    await initializeData();
  });

  async function initializeData() {
    isLoading = true;
    try {
      const monthsRes = await fetch(`${API_URL}/months`);
      if (monthsRes.ok) {
        months = await monthsRes.json() || [];
        if (months.length > 0 && !selectedMonth) {
          selectedMonth = months[0];
        }
      }
    } catch (e) {
      console.error('Failed to initialize leaderboard:', e);
    } finally {
      isInitialized = true;
      isLoading = false;
    }
  }

  async function fetchTickets() {
    try {
      const url = selectedMonth 
        ? `${API_URL}/tickets?month=${selectedMonth}`
        : `${API_URL}/tickets`;
      
      const res = await fetch(url);
      if (res.ok) {
        tickets = await res.json() || [];
      }
    } catch (e) {
      console.error('Error fetching tickets:', e);
    }
  }

  // Refetch when month changes
  $effect(() => {
    if (isInitialized && selectedMonth !== undefined) {
      fetchTickets();
    }
  });

  // Derived: Assignee aggregated statistics
  const assigneeStats = $derived.by(() => {
    const list = tickets;
    const counts: Record<string, { count: number; resolved: number; durationSum: number }> = {};

    list.forEach(t => {
      let name = t.Assignee;
      if (name) name = name.trim();
      if (!name || name === '-' || name === '') {
        name = 'Unassigned';
      }

      if (!counts[name]) {
        counts[name] = { count: 0, resolved: 0, durationSum: 0 };
      }

      counts[name].count++;

      const status = t.ticket_status.toLowerCase();
      const isResolved = status === 'closed' || status === 'resolved' || t.ResolvedAt || t.CompleteTime;
      if (isResolved) {
        counts[name].resolved++;
        const endVal = t.ResolvedAt || t.CompleteTime || t.ClosedDate;
        if (endVal) {
          const start = new Date(t.CreatedDate).getTime();
          const end = new Date(endVal).getTime();
          if (start && end && end >= start) {
            counts[name].durationSum += (end - start);
          }
        }
      }
    });

    let items = Object.entries(counts).map(([name, data]) => {
      const avgHrs = data.resolved > 0 
        ? (data.durationSum / data.resolved / (1000 * 60 * 60)).toFixed(1)
        : '-';
      const resolutionRate = data.count > 0 
        ? Math.round((data.resolved / data.count) * 100) 
        : 0;
      return {
        name,
        count: data.count,
        resolved: data.resolved,
        resolutionRate,
        avgHrs,
        durationSum: data.durationSum
      };
    });

    // Sort by count descending
    items.sort((a, b) => b.count - a.count);

    return items;
  });

  // Top 3 for the Podium
  const podiumData = $derived(assigneeStats.slice(0, 3));
  // Remaining list for the leaderboard rows
  const remainingData = $derived(assigneeStats.slice(3));

</script>

<svelte:window onclick={(e) => {
  // @ts-ignore
  if (isMonthDropdownOpen && !e.target.closest('.month-dropdown-wrapper')) {
    isMonthDropdownOpen = false;
  }
}} />

<svelte:head>
  <title>IAM Assignee Performance Leaderboard</title>
  <meta name="description" content="Detailed performance and efficiency rankings for IAM operations agents." />
</svelte:head>

<!-- Background glow components -->
<div class="fixed inset-0 -z-50 bg-slate-950 overflow-hidden">
  <div class="absolute -top-40 -left-40 w-96 h-96 bg-indigo-600/10 rounded-full blur-[128px]"></div>
  <div class="absolute top-1/3 -right-40 w-[500px] h-[500px] bg-cyan-600/10 rounded-full blur-[160px]"></div>
  <div class="absolute -bottom-40 left-1/3 w-[600px] h-[600px] bg-emerald-600/5 rounded-full blur-[180px]"></div>
  <!-- Grid -->
  <div class="absolute inset-0 bg-[linear-gradient(to_right,#0f172a_1px,transparent_1px),linear-gradient(to_bottom,#0f172a_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] opacity-30"></div>
</div>

<div class="min-h-screen flex flex-col">
  
  <!-- Header -->
  <header class="border-b border-slate-900/60 bg-slate-950/80 backdrop-blur-xl sticky top-0 z-30 px-6 py-4 flex items-center justify-between">
    <div class="flex items-center gap-3">
      <!-- Logo -->
      <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-cyan-500 flex items-center justify-center text-white shadow-md shadow-indigo-600/20">
        <svg xmlns="http://www.w3.org/2050/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
        </svg>
      </div>
      <div>
        <h1 class="text-base font-bold text-white tracking-tight">IAM Analytics</h1>
        <p class="text-[10px] text-slate-500 font-mono leading-none">V1.0.0 // Leaderboard</p>
      </div>
    </div>

    <nav class="hidden sm:flex items-center gap-1 bg-slate-900 border border-slate-850 p-1 rounded-xl text-xs font-semibold">
      <a href="/" class="px-3.5 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
        </svg>
        <span>Home</span>
      </a>
      <a href="/analysis" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Analytics</a>
      <a href="/leaderboard" class="px-4 py-2 rounded-lg bg-indigo-600 text-white shadow-md shadow-indigo-600/10 transition">Leaderboard</a>
      <a href="/incident" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Incidents</a>
    </nav>



    <!-- Month Dropdown filter inside header -->
    <div class="flex items-center gap-3">
      {#if months.length > 0}
        <div class="relative month-dropdown-wrapper inline-block">
          <button 
            onclick={() => isMonthDropdownOpen = !isMonthDropdownOpen}
            class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 hover:border-slate-700/80 hover:bg-slate-850 text-slate-200 text-xs font-semibold focus:outline-none transition shadow"
          >
            <svg class="w-3.5 h-3.5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
            </svg>
            <span>{selectedMonth || 'Select Month'}</span>
            <svg class="w-3.5 h-3.5 text-slate-500 transition-transform duration-200 {isMonthDropdownOpen ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7"/>
            </svg>
          </button>

          {#if isMonthDropdownOpen}
            <div 
              class="absolute right-0 mt-2 w-48 max-h-60 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-800 rounded-xl border border-slate-800 bg-slate-950/95 backdrop-blur-xl p-1.5 shadow-2xl z-40 space-y-0.5"
              transition:slide={{ duration: 150 }}
            >
              {#each months as month}
                <button 
                  onclick={() => { selectedMonth = month; isMonthDropdownOpen = false; }}
                  class="w-full text-left px-3 py-2 rounded-lg text-xs font-medium transition flex items-center justify-between
                    {selectedMonth === month 
                      ? 'bg-indigo-600 text-white font-semibold shadow-sm' 
                      : 'text-slate-400 hover:bg-slate-900/60 hover:text-white'}"
                >
                  <span>{month}</span>
                  {#if selectedMonth === month}
                    <svg class="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>
                  {/if}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </header>

  <main class="flex-1 max-w-5xl w-full mx-auto p-6 space-y-6">
    
    <!-- Title & Top Summary -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="space-y-1">
        <h2 class="text-2xl font-black tracking-tight text-white flex items-center gap-2">🏆 The IAM Hall of Legends</h2>
        <p class="text-xs text-slate-400 font-light">Where tickets go to die. Rankings are fueled by raw typing speed, caffeine, and pure willpower.</p>
      </div>
    </div>

    {#if isLoading}
      <div class="flex flex-col items-center justify-center py-24 space-y-4">
        <div class="relative w-12 h-12">
          <div class="absolute inset-0 border-4 border-indigo-500/20 rounded-full"></div>
          <div class="absolute inset-0 border-4 border-t-indigo-500 rounded-full animate-spin"></div>
        </div>
        <span class="text-xs font-mono text-slate-500">Compiling assignee statistics...</span>
      </div>
    {:else}
      
      <!-- Top 3 Podium Cards -->
      {#if podiumData.length > 0}
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 items-end pt-4">
          
          <!-- #2 Place Card (Left) -->
          {#if podiumData[1]}
            <div class="order-2 md:order-1 rounded-2xl border border-slate-900 bg-slate-950/20 backdrop-blur-xl p-5 space-y-4 shadow-xl text-center relative overflow-hidden flex flex-col justify-between min-h-[220px]">
              <div class="absolute -right-3 -top-3 w-12 h-12 bg-cyan-600/5 rounded-full blur-xl"></div>
              <div class="space-y-2">
                <span class="w-7 h-7 rounded-full bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 text-xs font-black font-mono inline-flex items-center justify-center">
                  #2
                </span>
                <span class="text-[9px] font-bold text-cyan-400/80 uppercase tracking-widest block font-mono">Keyboard Warrior ⚔️</span>
                <h3 class="text-sm font-bold text-slate-200 truncate">{podiumData[1].name}</h3>
                <p class="text-xs font-mono text-cyan-400 font-bold">{podiumData[1].count} resolved</p>
              </div>

              <div class="grid grid-cols-2 gap-2 pt-2 border-t border-slate-900/60 text-[10px] font-mono text-slate-400">
                <div class="bg-slate-950/40 p-2 rounded-xl">
                  <span class="text-[8px] text-slate-500 uppercase block tracking-wider leading-none mb-1">Speed</span>
                  <span class="font-bold text-slate-200">{podiumData[1].avgHrs} hrs</span>
                </div>
                <div class="bg-slate-950/40 p-2 rounded-xl">
                  <span class="text-[8px] text-slate-500 uppercase block tracking-wider leading-none mb-1">Success</span>
                  <span class="font-bold text-emerald-400">{podiumData[1].resolutionRate}%</span>
                </div>
              </div>
            </div>
          {/if}

          <!-- #1 Place Card (Center, Tallest) -->
          <div class="order-1 md:order-2 rounded-2xl border border-indigo-950 bg-indigo-950/10 backdrop-blur-xl p-6 space-y-4 shadow-xl text-center relative overflow-hidden flex flex-col justify-between min-h-[260px] md:-translate-y-2">
            <div class="absolute -right-3 -top-3 w-16 h-16 bg-indigo-500/10 rounded-full blur-xl"></div>
            <div class="absolute -left-3 -bottom-3 w-16 h-16 bg-violet-500/10 rounded-full blur-xl"></div>
            
            <div class="space-y-2 relative flex flex-col items-center">
              <!-- Beautiful Floating Gold Crown -->
              <svg class="w-7 h-7 text-amber-400 filter drop-shadow-[0_0_6px_rgba(245,158,11,0.6)] -mb-1 animate-bounce" style="animation-duration: 2.5s;" fill="currentColor" viewBox="0 0 24 24">
                <path d="M2 19h20v2H2v-2zm1-5l2.5-4.5 3 3.5 3.5-6.5 3.5 6.5 3-3.5 2.5 4.5H3z"/>
              </svg>
              <span class="w-9 h-9 rounded-full bg-gradient-to-r from-indigo-500 to-violet-600 text-white text-sm font-black font-mono inline-flex items-center justify-center shadow-[0_0_15px_rgba(99,102,241,0.4)]">
                #1
              </span>
              <span class="text-[9px] font-bold text-amber-400/90 uppercase tracking-widest block font-mono">Ticket Overlord 👑</span>
              <h3 class="text-base font-black text-white truncate w-full">{podiumData[0].name}</h3>
              <p class="text-xs font-mono text-indigo-400 font-bold">{podiumData[0].count} resolved</p>
            </div>

            <div class="grid grid-cols-2 gap-2 pt-2 border-t border-indigo-950/40 text-[10px] font-mono text-slate-400">
              <div class="bg-slate-950/60 p-2.5 rounded-xl border border-indigo-900/30">
                <span class="text-[8px] text-slate-500 uppercase block tracking-wider leading-none mb-1">Speed</span>
                <span class="font-bold text-slate-200">{podiumData[0].avgHrs} hrs</span>
              </div>
              <div class="bg-slate-950/60 p-2.5 rounded-xl border border-indigo-900/30">
                <span class="text-[8px] text-slate-500 uppercase block tracking-wider leading-none mb-1">Success</span>
                <span class="font-bold text-emerald-400">{podiumData[0].resolutionRate}%</span>
              </div>
            </div>
          </div>

          <!-- #3 Place Card (Right) -->
          {#if podiumData[2]}
            <div class="order-3 rounded-2xl border border-slate-900 bg-slate-950/20 backdrop-blur-xl p-5 space-y-4 shadow-xl text-center relative overflow-hidden flex flex-col justify-between min-h-[220px]">
              <div class="absolute -right-3 -top-3 w-12 h-12 bg-emerald-600/5 rounded-full blur-xl"></div>
              <div class="space-y-2">
                <span class="w-7 h-7 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-black font-mono inline-flex items-center justify-center">
                  #3
                </span>
                <span class="text-[9px] font-bold text-emerald-400/80 uppercase tracking-widest block font-mono">Coffee Consumer ☕</span>
                <h3 class="text-sm font-bold text-slate-200 truncate">{podiumData[2].name}</h3>
                <p class="text-xs font-mono text-emerald-400 font-bold">{podiumData[2].count} resolved</p>
              </div>

              <div class="grid grid-cols-2 gap-2 pt-2 border-t border-slate-900/60 text-[10px] font-mono text-slate-400">
                <div class="bg-slate-950/40 p-2 rounded-xl">
                  <span class="text-[8px] text-slate-500 uppercase block tracking-wider leading-none mb-1">Speed</span>
                  <span class="font-bold text-slate-200">{podiumData[2].avgHrs} hrs</span>
                </div>
                <div class="bg-slate-950/40 p-2 rounded-xl">
                  <span class="text-[8px] text-slate-500 uppercase block tracking-wider leading-none mb-1">Success</span>
                  <span class="font-bold text-emerald-400">{podiumData[2].resolutionRate}%</span>
                </div>
              </div>
            </div>
          {/if}

        </div>
      {/if}

      <!-- Assignee Rankings List -->
      <div class="rounded-3xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-6 space-y-5 shadow-xl">
        
        <div class="flex items-center justify-between border-b border-slate-900/60 pb-4">
          <div>
            <h3 class="text-sm font-bold text-slate-200">Active Agent Rankings</h3>
            <p class="text-xs text-slate-550">The Hall of Fame details.</p>
          </div>
        </div>

        <div class="space-y-2">
          {#if assigneeStats.length === 0}
            <div class="text-center py-12 text-slate-500 font-mono text-xs">
              No matching assignee records found.
            </div>
          {:else}
            <!-- Headers -->
            <div class="grid grid-cols-12 px-4 py-2 text-[9px] font-mono font-bold text-slate-500 uppercase tracking-wider">
              <div class="col-span-1">Rank</div>
              <div class="col-span-5 sm:col-span-4">Agent Name</div>
              <div class="col-span-3 text-right">Resolved Volume</div>
              <div class="col-span-3 sm:col-span-2 text-right">Avg Speed</div>
              <div class="hidden sm:block sm:col-span-2 text-right">Efficiency</div>
            </div>

            <!-- List rows -->
            {#each assigneeStats as agent, idx}
              <div class="grid grid-cols-12 items-center p-3 px-4 rounded-2xl border border-slate-900/60 bg-slate-950/20 hover:border-slate-800 hover:bg-slate-900/10 transition duration-150 gap-1">
                <!-- Rank -->
                <div class="col-span-1 text-xs font-mono font-bold text-slate-400">
                  #{idx + 1}
                </div>

                <!-- Agent Name -->
                <div class="col-span-5 sm:col-span-4 font-bold text-slate-200 truncate">
                  {agent.name}
                </div>

                <!-- Volume -->
                <div class="col-span-3 text-right text-xs font-mono text-slate-300 font-bold">
                  {agent.count} <span class="text-[10px] text-slate-550 font-normal font-sans">tickets</span>
                </div>

                <!-- Speed -->
                <div class="col-span-3 sm:col-span-2 text-right text-xs font-mono text-cyan-400 font-bold">
                  {agent.avgHrs} <span class="text-[10px] text-slate-550 font-normal font-sans">hrs</span>
                </div>

                <!-- Efficiency rate progress bar -->
                <div class="hidden sm:flex col-span-2 items-center gap-3 justify-end">
                  <div class="text-right w-12 font-mono">
                    <span class="text-emerald-400 font-bold text-xs">{agent.resolutionRate}%</span>
                  </div>
                  <div class="h-1.5 bg-slate-900 rounded-full overflow-hidden w-16">
                    <div 
                      class="h-full bg-emerald-500 rounded-full" 
                      style="width: {agent.resolutionRate}%"
                    ></div>
                  </div>
                </div>
              </div>
            {/each}
          {/if}
        </div>

        <div class="text-[9px] font-mono text-slate-600 flex items-center justify-between pt-4 border-t border-slate-900/60">
          <span>Displaying {assigneeStats.length} active agent profiles</span>
          <span>Source: IAM Tickets Database</span>
        </div>

      </div>

    {/if}
  </main>
</div>
