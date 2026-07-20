<script lang="ts">
  import { onMount } from 'svelte';
  import { fade, fly, slide } from 'svelte/transition';

  // API base URL
  const API_URL = 'http://localhost:8081';

  interface Ticket {
    ticket_id: string;
    ticket_type: string;
    subject: string;
    site?: string;
    site_group?: string;
    region_site?: string;
    domain_group: string;
    company: string;
    country: string;
    priority: string;
    ticket_status: string;
    customer_name: string;
    ProductT1?: string;
    ProductT2?: string;
    ProductT3?: string;
    GroupAssignee?: string;
    Assignee?: string;
    CreatedDate: string;
    ResolvedAt?: string;
    ClosedDate?: string;
    CompleteTime?: string;
    DetailDescription?: string;
    SourceFile: string;
    uploaded_at: string;
  }

  // Reactive state using Runes
  let isLoading = $state(true);
  let tickets = $state<Ticket[]>([]);
  let months = $state<string[]>([]);
  let selectedMonth = $state<string>('');
  let isMonthDropdownOpen = $state(false);

  // Active Analysis Tier Tab
  let activeTier = $state<'T1' | 'T2' | 'T3'>('T1');

  // Interactivity States
  let hoveredCategoryName = $state<string | null>(null);
  let selectedCategoryName = $state<string | null>(null);
  let categorySearch = $state<string>('');
  let categorySort = $state<'count_desc' | 'count_asc' | 'name_asc'>('count_desc');

  // Tooltip States
  let tooltipText = $state<string | null>(null);
  let tooltipX = $state(0);
  let tooltipY = $state(0);

  function showTooltip(e: MouseEvent, text: string) {
    tooltipText = text;
    tooltipX = e.clientX + 12;
    tooltipY = e.clientY + 12;
  }

  function moveTooltip(e: MouseEvent) {
    tooltipX = e.clientX + 12;
    tooltipY = e.clientY + 12;
  }

  function hideTooltip() {
    tooltipText = null;
  }

  // SVG Geometry Config for a Minimal Premium Donut Chart
  const r = 45;
  const c = 2 * Math.PI * r; // ~282.743
  const strokeWidth = 12; // Donut thickness

  // Premium Curated Color Palette
  const colors = [
    '#6366f1', // Indigo
    '#06b6d4', // Cyan
    '#10b981', // Emerald
    '#f59e0b', // Amber
    '#d946ef', // Fuchsia
    '#ec4899', // Pink
    '#8b5cf6', // Violet
    '#3b82f6', // Blue
    '#065f46', // Dark Green
    '#84cc16', // Lime
  ];
  const uncategorizedColor = '#475569'; // Slate 600
  const otherColor = '#64748b'; // Slate 500

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
      console.error('Failed to initialize analysis:', e);
    } finally {
      isInitialized = true;
      isLoading = false;
    }
  }

  let selectedT1Category = $state<string | null>(null);
  let selectedT2Category = $state<string | null>(null);

  async function fetchTickets() {
    try {
      const url = selectedMonth 
        ? `${API_URL}/tickets?month=${selectedMonth}`
        : `${API_URL}/tickets`;
      
      const res = await fetch(url);
      if (res.ok) {
        tickets = await res.json() || [];
        selectedCategoryName = null;
        hoveredCategoryName = null;
        selectedT1Category = null;
        selectedT2Category = null;
      }
    } catch (e) {
      console.error('Error fetching tickets:', e);
    }
  }

  // Refetch when month changes, only after initial month lookup
  $effect(() => {
    if (isInitialized && selectedMonth !== undefined) {
      fetchTickets();
    }
  });

  // Reset category search and selections when changing tiers
  $effect(() => {
    if (activeTier) {
      categorySearch = '';
      selectedCategoryName = null;
      hoveredCategoryName = null;
    }
  });

  // Sync selected categories down the hierarchy when user changes selections
  $effect(() => {
    if (activeTier === 'T1' && selectedCategoryName && selectedCategoryName !== 'Uncategorized' && selectedCategoryName !== 'Other') {
      selectedT1Category = selectedCategoryName;
    }
  });

  $effect(() => {
    if (activeTier === 'T2' && selectedCategoryName && selectedCategoryName !== 'Uncategorized' && selectedCategoryName !== 'Other') {
      selectedT2Category = selectedCategoryName;
    }
  });

  // --- Description Filter Engine ---
  let descriptionSearchInput = $state<string>('');

  // --- Incident Tickets Filtering ---
  let showIncidentTickets = $state(true);

  // Group tickets in current dataset by day to identify incident days (>= 50 tickets)
  const incidentDates = $derived.by(() => {
    const counts: Record<string, number> = {};
    tickets.forEach(t => {
      const d = new Date(t.CreatedDate);
      if (!isNaN(d.getTime())) {
        const key = d.toDateString();
        counts[key] = (counts[key] || 0) + 1;
      }
    });
    return counts;
  });

  // Count of incident tickets in the current month/dataset
  const incidentTicketsCount = $derived(
    tickets.filter(t => {
      const d = new Date(t.CreatedDate);
      if (!isNaN(d.getTime())) {
        const key = d.toDateString();
        return (incidentDates[key] || 0) >= 50;
      }
      return false;
    }).length
  );

  // Derived: Filter tickets by custom regex search query in Subject or DetailDescription and incident ticket status
  const filteredTickets = $derived.by(() => {
    let result = tickets;

    // Filter out incident tickets if toggle is off
    if (!showIncidentTickets) {
      result = result.filter(t => {
        const d = new Date(t.CreatedDate);
        if (!isNaN(d.getTime())) {
          const key = d.toDateString();
          return (incidentDates[key] || 0) < 50;
        }
        return true;
      });
    }

    if (descriptionSearchInput.trim()) {
      try {
        const regex = new RegExp(descriptionSearchInput.trim(), 'i');
        result = result.filter(t => regex.test(t.subject || '') || regex.test(t.DetailDescription || ''));
      } catch (e) {
        // Silently ignore invalid regex during typing
      }
    }

    return result;
  });
  // --- End of Description Filter Engine ---

  // core categorization data extractor
  const rawCategoryStats = $derived.by(() => {
    let list = filteredTickets;

    // Drill-down filtering based on previous tiers
    if (activeTier === 'T2') {
      if (selectedT1Category) {
        list = list.filter(t => {
          let val = t.ProductT1;
          if (val) val = val.trim();
          const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
          return catName === selectedT1Category;
        });
      }
    } else if (activeTier === 'T3') {
      if (selectedT1Category) {
        list = list.filter(t => {
          let val = t.ProductT1;
          if (val) val = val.trim();
          const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
          return catName === selectedT1Category;
        });
      }
      if (selectedT2Category) {
        list = list.filter(t => {
          let val = t.ProductT2;
          if (val) val = val.trim();
          const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
          return catName === selectedT2Category;
        });
      }
    }

    const counts: Record<string, number> = {};
    let totalValid = 0;
    const field = activeTier === 'T1' ? 'ProductT1' : activeTier === 'T2' ? 'ProductT2' : 'ProductT3';

    list.forEach(t => {
      let val = t[field];
      if (val) val = val.trim();
      const name = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
      counts[name] = (counts[name] || 0) + 1;
      totalValid++;
    });

    return { counts, totalValid };
  });

  // Slices generation with sorting, search, and capping (Other category)
  const chartData = $derived.by(() => {
    const { counts, totalValid } = rawCategoryStats;
    if (totalValid === 0) return [];

    // Sort descending by default for slicing top elements
    const sorted = Object.entries(counts).sort((a, b) => {
      if (a[0] === 'Uncategorized') return 1;
      if (b[0] === 'Uncategorized') return -1;
      return b[1] - a[1];
    });

    const maxSlices = 8;
    let finalItems: [string, number][] = [];
    let otherCount = 0;

    sorted.forEach(([name, count], index) => {
      if (sorted.length > maxSlices && index >= maxSlices - 1 && name !== 'Uncategorized') {
        otherCount += count;
      } else {
        finalItems.push([name, count]);
      }
    });

    if (otherCount > 0) {
      const uncategIndex = finalItems.findIndex(([name]) => name === 'Uncategorized');
      if (uncategIndex !== -1) {
        const [uncateg] = finalItems.splice(uncategIndex, 1);
        finalItems.push(['Other', otherCount]);
        finalItems.push(uncateg);
      } else {
        finalItems.push(['Other', otherCount]);
      }
    }

    // Map colors, percentage, and compute SVG dash array/offsets
    let accumulatedLength = 0;
    const mapped = finalItems.map(([name, count], idx) => {
      const percentage = (count / totalValid) * 100;
      const length = (count / totalValid) * c;
      
      let color = colors[idx % colors.length];
      if (name === 'Uncategorized') color = uncategorizedColor;
      else if (name === 'Other') color = otherColor;

      // Slice gap: subtract 1 unit to create clean hairline gaps between solid wedges
      const dashLength = Math.max(0.2, length - 1);

      // Midpoint calculations for placing text labels directly inside the wedges
      const midLength = accumulatedLength + length / 2;
      const midAngle = -Math.PI / 2 + (midLength / c) * 2 * Math.PI;
      const textRadius = 18; // Radius at which to draw percentages (well inside outer edge)
      const labelX = 65 + textRadius * Math.cos(midAngle);
      const labelY = 65 + textRadius * Math.sin(midAngle);

      const item = {
        name,
        count,
        percentage,
        color,
        strokeDashArray: `${dashLength} ${c}`,
        strokeDashOffset: -accumulatedLength,
        labelX,
        labelY,
      };
      accumulatedLength += length;
      return item;
    });

    return mapped;
  });

  // Filtered List View for the Legend Grid (Includes sorting + text search)
  const listData = $derived.by(() => {
    const { counts, totalValid } = rawCategoryStats;
    if (totalValid === 0) return [];

    let items = Object.entries(counts).map(([name, count]) => {
      const percentage = (count / totalValid) * 100;
      
      const chartMatch = chartData.find(c => c.name === name);
      let color = chartMatch?.color;
      
      if (!color) {
        if (name === 'Uncategorized') color = uncategorizedColor;
        else color = otherColor;
      }

      return { name, count, percentage, color };
    });

    // Apply Search Filter
    if (categorySearch.trim()) {
      const query = categorySearch.toLowerCase();
      items = items.filter(item => item.name.toLowerCase().includes(query));
    }

    // Apply Sorting Options
    items.sort((a, b) => {
      if (categorySort === 'count_desc') {
        return b.count - a.count;
      } else if (categorySort === 'count_asc') {
        return a.count - b.count;
      } else {
        return a.name.localeCompare(b.name);
      }
    });

    return items;
  });

  // Highlighted Top 3 Categories for Instant Insight
  const topInsightCategories = $derived.by(() => {
    const { counts, totalValid } = rawCategoryStats;
    if (totalValid === 0) return [];

    // Sort by count descending
    const sorted = Object.entries(counts)
      .map(([name, count]) => ({
        name,
        count,
        percentage: (count / totalValid) * 100,
      }))
      .sort((a, b) => b.count - a.count);

    // Map color to top 3
    return sorted.slice(0, 3).map((item, idx) => {
      const chartMatch = chartData.find(c => c.name === item.name);
      const color = chartMatch?.color || colors[idx] || otherColor;
      return { ...item, color, rank: idx + 1 };
    });
  });

  // Selected Category Information (For Details Panel)
  const activeCategoryInfo = $derived.by(() => {
    const targetName = selectedCategoryName;
    if (!targetName) {
      return {
        name: `All Product ${activeTier} Scope`,
        count: rawCategoryStats.totalValid,
        percentage: 100,
        color: '#6366f1'
      };
    }

    const match = listData.find(item => item.name === targetName) || 
                  chartData.find(slice => slice.name === targetName);
    
    return match || null;
  });

  // Derived state for the premium donut chart center readout
  const centerDisplay = $derived.by(() => {
    const activeName = hoveredCategoryName || selectedCategoryName;
    if (activeName) {
      const match = listData.find(item => item.name === activeName) || 
                    chartData.find(slice => slice.name === activeName);
      if (match) {
        return {
          label: match.name.toUpperCase(),
          value: `${match.percentage.toFixed(1)}%`,
          subLabel: `${match.count.toLocaleString()} tickets`,
          color: match.color || '#6366f1'
        };
      }
    }
    return {
      label: 'TOTAL TICKETS',
      value: rawCategoryStats.totalValid.toLocaleString(),
      subLabel: 'Active Scope',
      color: '#6366f1'
    };
  });

  let previewSearchQuery = $state('');
  let detailTicket = $state<Ticket | null>(null);

  // Extract all tickets belonging to the selected category (or all scope tickets if none selected)
  const categoryTickets = $derived.by(() => {
    const targetName = selectedCategoryName;

    const field = activeTier === 'T1' ? 'ProductT1' : activeTier === 'T2' ? 'ProductT2' : 'ProductT3';

    let list = filteredTickets;
    if (activeTier === 'T2') {
      if (selectedT1Category) {
        list = list.filter(t => {
          let val = t.ProductT1;
          if (val) val = val.trim();
          const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
          return catName === selectedT1Category;
        });
      }
    } else if (activeTier === 'T3') {
      if (selectedT1Category) {
        list = list.filter(t => {
          let val = t.ProductT1;
          if (val) val = val.trim();
          const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
          return catName === selectedT1Category;
        });
      }
      if (selectedT2Category) {
        list = list.filter(t => {
          let val = t.ProductT2;
          if (val) val = val.trim();
          const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;
          return catName === selectedT2Category;
        });
      }
    }

    if (!targetName) {
      return list;
    }

    return list.filter(t => {
      let val = t[field];
      if (val) val = val.trim();
      const catName = (!val || val === '-' || val === '') ? 'Uncategorized' : val;

      if (targetName === 'Other') {
        const topNames = chartData.map(c => c.name).filter(n => n !== 'Other' && n !== 'Uncategorized');
        return catName !== 'Uncategorized' && !topNames.includes(catName);
      }
      return catName === targetName;
    });
  });

  // Filter category tickets based on search query (Ticket ID or Subject)
  const filteredPreviewTickets = $derived.by(() => {
    let result = categoryTickets;
    if (previewSearchQuery.trim()) {
      const query = previewSearchQuery.toLowerCase().trim();
      result = result.filter(t => 
        t.ticket_id.toLowerCase().includes(query) ||
        t.subject.toLowerCase().includes(query)
      );
    }
    return result;
  });

  // Reset category search when selected category changes
  $effect(() => {
    if (selectedCategoryName) {
      previewSearchQuery = '';
    }
  });

  // Date formatter helper
  function formatDate(dateStr: string) {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

</script>

<svelte:window onclick={(e) => {
  // @ts-ignore
  if (isMonthDropdownOpen && !e.target.closest('.month-dropdown-wrapper')) {
    isMonthDropdownOpen = false;
  }
}} />

<svelte:head>
  <title>Product Categorization Analysis</title>
  <meta name="description" content="Ultimate interactive dashboard for instant comprehension of product tiers." />
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
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
        </svg>
      </div>
      <div>
        <h1 class="text-base font-bold text-white tracking-tight">IAM Analytics</h1>
        <p class="text-[10px] text-slate-500 font-mono leading-none">V1.0.0 // Analytics</p>
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
      <a href="/analysis" class="px-4 py-2 rounded-lg bg-indigo-600 text-white shadow-md shadow-indigo-600/10 transition">Analytics</a>
      <a href="/leaderboard" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Leaderboard</a>
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

  <main class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">
    
    <!-- Title & Top Tier Navigation -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="space-y-1">
        <h2 class="text-xl font-bold tracking-tight text-white">Product Categorization Hub</h2>
        <p class="text-xs text-slate-400 font-light">Interactive analysis and segmented solid pie charts for Product Tiers 1, 2, and 3.</p>
      </div>

      <!-- Tier Tabs Selector -->
      <div class="flex bg-slate-950 border border-slate-900 p-1 rounded-xl text-xs font-semibold self-start md:self-auto shadow-inner">
        {#each ['T1', 'T2', 'T3'] as tier}
          <button 
            onclick={() => activeTier = tier as any}
            class="px-5 py-2.5 rounded-lg transition duration-200 border text-[11px] uppercase tracking-wider
              {activeTier === tier 
                ? 'bg-gradient-to-r from-indigo-500 to-violet-600 text-white font-bold border-indigo-400/30 shadow-[0_0_15px_rgba(99,102,241,0.35)]' 
                : 'text-slate-400 border-transparent hover:text-white hover:bg-slate-900/60'}"
          >
            Product {tier}
          </button>
        {/each}
      </div>
    </div>

    <!-- Tier hierarchy breadcrumb indicator -->
    {#if activeTier !== 'T1'}
      <div class="flex flex-wrap items-center gap-2.5 text-xs font-mono px-5 py-3.5 bg-gradient-to-r from-slate-900 to-indigo-950/15 border-l-4 border-l-indigo-500 border-y border-r border-slate-900 rounded-xl shadow-lg">
        <span class="font-bold text-slate-400 uppercase tracking-widest text-[9px] mr-1 flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse"></span>
          Active Scope:
        </span>
        <span class="text-indigo-300 font-bold px-2.5 py-1 rounded-lg bg-indigo-950/40 border border-indigo-900/40 shadow-sm">{selectedT1Category || 'All T1'}</span>
        {#if activeTier === 'T3'}
          <svg class="w-4 h-4 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7"/></svg>
          <span class="text-indigo-300 font-bold px-2.5 py-1 rounded-lg bg-indigo-950/40 border border-indigo-900/40 shadow-sm">{selectedT2Category || 'All T2'}</span>
        {/if}
      </div>
    {/if}

    <!-- Description Search Bar (Regex) -->
    <div class="rounded-2xl border border-slate-900 bg-slate-950/40 p-4">
      <div class="relative w-full">
        <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-4 h-4 text-slate-400">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.637 10.637z" />
          </svg>
        </span>
        <input
          type="text"
          bind:value={descriptionSearchInput}
          placeholder="Filter by description keyword or regex (e.g. ต่ออายุ, error.*failed, request.*access)..."
          class="w-full bg-slate-900 border border-slate-800 placeholder-slate-550 text-slate-200 rounded-xl pl-10 pr-4 py-2.5 text-xs focus:outline-none focus:border-indigo-500/80 focus:ring-1 focus:ring-indigo-500/20 shadow-inner"
        />
      </div>
    </div>

    {#if isLoading}
      <!-- Loader -->
      <div class="flex flex-col items-center justify-center py-32 space-y-4">
        <div class="relative w-16 h-16">
          <div class="absolute inset-0 border-4 border-indigo-500/20 rounded-full"></div>
          <div class="absolute inset-0 border-4 border-t-indigo-500 rounded-full animate-spin"></div>
        </div>
        <p class="text-sm text-slate-400 font-mono">Aggregating product metadata...</p>
      </div>

    {:else if tickets.length === 0}
      <!-- Empty State -->
      <div class="border border-slate-900 bg-slate-950/40 rounded-3xl p-16 text-center max-w-xl mx-auto space-y-4 backdrop-blur-xl">
        <div class="w-12 h-12 bg-slate-900 border border-slate-800 rounded-xl flex items-center justify-center mx-auto text-slate-500">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
          </svg>
        </div>
        <h3 class="text-lg font-bold text-slate-200">No data found</h3>
        <p class="text-xs text-slate-500">There are no ticket records available in the database. Go back to the Register page to sync your spreadsheet.</p>
        <a href="/" class="inline-block px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-xs font-semibold shadow transition">Go to Register</a>
      </div>

    {:else}
      <!-- 3-Second Insight Header Panel -->
      <section class="grid grid-cols-1 md:grid-cols-3 gap-4" in:fade>
        {#each topInsightCategories as category, idx}
          <div 
            class="relative rounded-2xl border bg-slate-950/50 backdrop-blur-md p-4 transition-all duration-300 cursor-pointer flex flex-col justify-between overflow-hidden group
              {selectedCategoryName === category.name 
                ? 'border-indigo-500 bg-indigo-950/10 shadow-lg' 
                : 'border-slate-900 hover:border-slate-800'}"
            onclick={() => selectedCategoryName = category.name}
            role="button"
            tabindex="0"
          >
            <!-- Decorative color accent bar -->
            <div class="absolute left-0 top-0 bottom-0 w-1" style="background-color: {category.color}"></div>

            <div class="space-y-1 pl-2">
              <div class="flex items-center justify-between text-[10px] font-mono text-slate-500">
                <span>RANK #{category.rank}</span>
                <span class="font-bold" style="color: {category.color}">
                  {idx === 0 ? 'PRIMARY' : idx === 1 ? 'SECONDARY' : 'TERTIARY'} VOLUME
                </span>
              </div>
              <h4 class="text-xs font-bold text-slate-200 truncate group-hover:text-white transition" title={category.name}>
                {category.name}
              </h4>
            </div>

            <div class="flex items-baseline justify-between pt-4 pl-2">
              <span class="text-2xl font-black text-white leading-none tracking-tight">{category.percentage.toFixed(1)}%</span>
              <span class="text-[10px] font-mono text-slate-500">{category.count} tickets</span>
            </div>
          </div>
        {/each}
      </section>

      <!-- Two-column split interface -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
        
        <!-- Left Side: Solid Pie Chart & Detail Panel (Grid Span 5) -->
        <div class="lg:col-span-5 flex flex-col gap-6">
          
          <!-- Solid Pie Chart Card -->
          <div class="rounded-3xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-6 flex flex-col items-center justify-center min-h-[350px] shadow-xl relative overflow-hidden group {!activeCategoryInfo ? 'flex-1 h-full' : ''}">
            <div class="absolute -right-16 -top-16 w-36 h-36 bg-indigo-500/5 rounded-full blur-2xl group-hover:bg-indigo-500/10 transition duration-500"></div>
            
            <!-- Incident Ticket Toggle (Small Top Right) -->
            <button
              onclick={() => showIncidentTickets = !showIncidentTickets}
              class="absolute top-5 right-5 z-20 flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border transition-all duration-300 text-[10px] font-bold tracking-wide shadow-md cursor-pointer select-none
                {showIncidentTickets 
                  ? 'bg-indigo-600/10 border-indigo-500/30 text-indigo-300 hover:bg-indigo-600/20' 
                  : 'bg-slate-900/80 border-slate-800 text-slate-400 hover:border-slate-700'}"
            >
              <span class="relative flex h-1.5 w-1.5">
                {#if showIncidentTickets}
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-indigo-500"></span>
                {:else}
                  <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-slate-650"></span>
                {/if}
              </span>
              <span>Incidents</span>
              {#if incidentTicketsCount > 0}
                <span class="font-mono text-[9px] px-1 bg-rose-500/25 border border-rose-500/30 text-rose-300 rounded">
                  {incidentTicketsCount}
                </span>
              {/if}
              <span class="ml-0.5 opacity-80">{showIncidentTickets ? 'ON' : 'OFF'}</span>
            </button>

            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-450 self-start mb-6 flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-indigo-500 animate-pulse"></span>
              {activeTier === 'T1' ? 'Tier 1' : activeTier === 'T2' ? 'Tier 2' : 'Tier 3'} Pie Chart
            </h3>

            <!-- Donut Graphic Container -->
            <div class="relative flex items-center justify-center">
              <svg width="240" height="240" viewBox="0 0 130 130" class="overflow-visible select-none">
                <!-- Drop shadow definitions and gradients for extra premium look -->
                <defs>
                  <filter id="donut-glow" x="-20%" y="-20%" width="140%" height="140%">
                    <feDropShadow dx="0" dy="4" stdDeviation="6" flood-color="#000000" flood-opacity="0.5"/>
                  </filter>
                  <radialGradient id="center-glow" cx="50%" cy="50%" r="50%">
                    <stop offset="0%" stop-color="#6366f1" stop-opacity="0.08" />
                    <stop offset="100%" stop-color="#6366f1" stop-opacity="0" />
                  </radialGradient>
                </defs>

                <!-- Ambient center glow -->
                <circle cx="65" cy="65" r="38" fill="url(#center-glow)" />

                <!-- Thin outer track and inner track borders for high-fidelity UI -->
                <circle cx="65" cy="65" r="51.5" fill="none" stroke="#1e293b" stroke-width="0.75" opacity="0.5" />
                <circle cx="65" cy="65" r="38.5" fill="none" stroke="#1e293b" stroke-width="0.75" opacity="0.5" />

                <!-- Faint background donut track -->
                <circle cx="65" cy="65" r={r} fill="none" stroke="#0f172a" stroke-width={strokeWidth} opacity="0.6" />
                
                {#each chartData as slice, idx}
                  <!-- Donut segment -->
                  <circle 
                    cx="65" 
                    cy="65" 
                    r={r} 
                    fill="none" 
                    stroke={slice.color} 
                    stroke-width={strokeWidth} 
                    stroke-dasharray={slice.strokeDashArray}
                    stroke-dashoffset={slice.strokeDashOffset}
                    class="cursor-pointer transition-all duration-300 origin-center"
                    style="
                      opacity: (hoveredCategoryName === null && selectedCategoryName === null) || hoveredCategoryName === slice.name || selectedCategoryName === slice.name ? 1 : 0.25;
                      stroke-width: {hoveredCategoryName === slice.name || selectedCategoryName === slice.name ? strokeWidth + 2.5 : strokeWidth}px;
                      transform: rotate(-90deg);
                      transform-origin: 65px 65px;
                      filter: {hoveredCategoryName === slice.name || selectedCategoryName === slice.name ? 'drop-shadow(0px 0px 4px ' + slice.color + '40)' : 'none'};
                    "
                    onmouseenter={(e) => { hoveredCategoryName = slice.name; showTooltip(e, slice.name); }}
                    onmousemove={moveTooltip}
                    onmouseleave={() => { hoveredCategoryName = null; hideTooltip(); }}
                    onclick={() => selectedCategoryName = selectedCategoryName === slice.name ? null : slice.name}
                    role="button"
                    tabindex="0"
                  />
                {/each}

                <!-- Center Readout Texts wrapped in foreignObject for autowrapping & alignment -->
                <foreignObject x="29" y="29" width="72" height="72">
                  <div class="w-full h-full flex flex-col items-center justify-center text-center select-none pointer-events-none leading-[1.1] p-1 font-mono">
                    <span 
                      class="text-[5px] font-bold uppercase transition-all duration-300 line-clamp-2 break-words overflow-hidden" 
                      style="color: {centerDisplay.color}"
                    >
                      {centerDisplay.label}
                    </span>
                    <span class="text-xs font-black tracking-tight text-white mt-1.5 drop-shadow-md">
                      {centerDisplay.value}
                    </span>
                    <span class="text-[5px] text-slate-500 font-semibold mt-0.5">
                      {centerDisplay.subLabel}
                    </span>
                  </div>
                </foreignObject>
              </svg>
            </div>

            <!-- Dynamic Bottom Info Label -->
            <p class="text-[9px] text-slate-500 mt-6 font-mono text-center flex items-center gap-1.5 justify-center">
              <span class="w-1.5 h-1.5 rounded-full bg-indigo-500 animate-pulse"></span>
              Click slices to lock selection & filter tickets
            </p>
          </div>

          <!-- Drill Down Category Focus Panel -->
          {#if activeCategoryInfo}
            <div class="rounded-3xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-6 space-y-4 shadow-xl" transition:fade>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 truncate max-w-[60%]">
                  <span class="w-3 h-3 rounded-full shrink-0" style="background-color: {activeCategoryInfo.color}"></span>
                  <h3 class="text-sm font-bold text-slate-200 truncate" title={activeCategoryInfo.name}>{activeCategoryInfo.name}</h3>
                </div>
                <div class="flex items-center gap-2">
                  {#if selectedCategoryName}
                    <button 
                      onclick={() => selectedCategoryName = null}
                      class="text-[10px] font-bold text-indigo-400 hover:text-indigo-350 underline transition"
                    >
                      Clear Selection
                    </button>
                  {/if}
                  <span class="text-xs font-mono font-bold bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 px-2 py-0.5 rounded">
                    {activeCategoryInfo.percentage.toFixed(1)}%
                  </span>
                </div>
              </div>

              <!-- Metrics -->
              <div class="grid grid-cols-2 gap-3 pt-2">
                <div class="bg-slate-950 border border-slate-900/80 p-3 rounded-2xl">
                  <span class="text-[9px] font-bold text-slate-500 uppercase tracking-wider block">Ticket Count</span>
                  <span class="text-lg font-black text-white font-mono">{activeCategoryInfo.count}</span>
                </div>
                <div class="bg-slate-950 border border-slate-900/80 p-3 rounded-2xl">
                  <span class="text-[9px] font-bold text-slate-500 uppercase tracking-wider block">Index Status</span>
                  <span class="text-xs font-semibold text-emerald-400 font-mono block mt-1">✔ Online</span>
                </div>
              </div>

              <!-- Category Tickets Preview list -->
              <div class="space-y-3 pt-2">
                <span class="text-[10px] font-bold text-slate-455 uppercase tracking-wider block">Tickets in Category ({filteredPreviewTickets.length})</span>
                
                <!-- Ticket ID / Subject Search Box -->
                <div class="relative">
                  <span class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-slate-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-3.5 h-3.5">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.637 10.637z" />
                    </svg>
                  </span>
                  <input 
                    type="text" 
                    bind:value={previewSearchQuery}
                    placeholder="Search Ticket ID or subject..." 
                    class="w-full bg-slate-950 border border-slate-900 placeholder-slate-600 text-slate-350 rounded-xl pl-8 pr-3 py-1.5 text-[10px] focus:outline-none focus:border-indigo-500/80"
                  />
                </div>

                {#if filteredPreviewTickets.length === 0}
                  <p class="text-[10px] font-mono text-slate-500 italic py-2">No matching tickets found.</p>
                {:else}
                  <div class="space-y-1.5 max-h-[160px] overflow-y-auto scrollbar-thin scrollbar-thumb-slate-900 pr-1">
                    {#each filteredPreviewTickets as t}
                      <div 
                        onclick={() => detailTicket = t}
                        class="p-2 rounded-xl bg-slate-950/60 border border-slate-900/60 flex items-center justify-between text-[11px] hover:border-slate-800 hover:bg-slate-900/40 transition cursor-pointer"
                        role="button"
                        tabindex="0"
                      >
                        <span class="font-mono text-slate-400 truncate max-w-[70%]" title={t.subject}>{t.subject}</span>
                        <span class="font-mono text-indigo-400 font-bold shrink-0">{t.ticket_id}</span>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            </div>
          {/if}

        </div>

        <!-- Right Side: Highly readable list view with progress bars (Grid Span 7) -->
        <div class="lg:col-span-7 rounded-3xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-6 flex flex-col justify-between shadow-xl self-start">
          
          <div class="space-y-4">
            
            <!-- List Header controls -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-900/60 pb-4">
              <div>
                <h3 class="text-sm font-bold text-slate-200">Category breakdown list</h3>
                <p class="text-xs text-slate-555 mt-0.5">Highly readable details with percentage meters.</p>
              </div>

              <!-- Sorting Pills -->
              <div class="flex bg-slate-900 border border-slate-850 p-0.5 rounded-xl text-[10px] font-bold self-start sm:self-auto">
                <button 
                  onclick={() => categorySort = 'count_desc'}
                  class="px-2.5 py-1.5 rounded-lg transition {categorySort === 'count_desc' ? 'bg-slate-850 text-white' : 'text-slate-555 hover:text-slate-355'}"
                >
                  Most Count
                </button>
                <button 
                  onclick={() => categorySort = 'count_asc'}
                  class="px-2.5 py-1.5 rounded-lg transition {categorySort === 'count_asc' ? 'bg-slate-850 text-white' : 'text-slate-555 hover:text-slate-355'}"
                >
                  Least Count
                </button>
                <button 
                  onclick={() => categorySort = 'name_asc'}
                  class="px-2.5 py-1.5 rounded-lg transition {categorySort === 'name_asc' ? 'bg-slate-850 text-white' : 'text-slate-555 hover:text-slate-355'}"
                >
                  Alphabetical
                </button>
              </div>
            </div>

            <!-- Search bar within Category List -->
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-4 h-4">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.637 10.637z" />
                </svg>
              </span>
              <input 
                type="text" 
                bind:value={categorySearch}
                placeholder="Search {activeTier === 'T1' ? 'Tier 1' : activeTier === 'T2' ? 'Tier 2' : 'Tier 3'} categories..." 
                class="w-full bg-slate-900 border border-slate-855 placeholder-slate-555 text-slate-200 rounded-xl pl-9 pr-4 py-2.5 text-xs focus:outline-none focus:border-indigo-500/80 focus:ring-1 focus:ring-indigo-500/20"
              />
            </div>

            <!-- Scrollable list of categories -->
            <div class="space-y-1 h-[576px] overflow-y-auto pr-2 scrollbar-thin scrollbar-thumb-slate-900">
              {#if listData.length === 0}
                <div class="text-center text-slate-555 font-mono text-xs py-8">
                  No matching category records found.
                </div>
              {:else}
                {#each listData as item}
                  <div 
                    class="p-3 rounded-2xl border transition duration-150 cursor-pointer flex flex-col gap-2
                      {selectedCategoryName === item.name 
                        ? 'border-indigo-500 bg-indigo-950/15 shadow-sm shadow-indigo-650/5' 
                        : hoveredCategoryName === item.name 
                          ? 'border-slate-800 bg-slate-900/40' 
                          : 'border-slate-900 bg-transparent hover:border-slate-800/80'}"
                    onmouseenter={(e) => { hoveredCategoryName = item.name; showTooltip(e, item.name); }}
                    onmousemove={moveTooltip}
                    onmouseleave={() => { hoveredCategoryName = null; hideTooltip(); }}
                    onclick={() => selectedCategoryName = selectedCategoryName === item.name ? null : item.name}
                    role="button"
                    tabindex="0"
                  >
                    <!-- Detail row -->
                    <div class="flex items-center justify-between text-xs font-semibold">
                      <div class="flex items-center gap-2.5 truncate max-w-[70%]">
                        <span class="w-2.5 h-2.5 rounded-full shrink-0" style="background-color: {item.color}"></span>
                        <span class="text-slate-250 truncate" title={item.name}>{item.name}</span>
                      </div>
                      <div class="text-right flex items-center gap-2">
                        <span class="font-mono text-slate-200">{item.percentage.toFixed(1)}%</span>
                        <span class="text-[10px] font-mono text-slate-500">({item.count} tickets)</span>
                      </div>
                    </div>

                    <!-- Progress bar breakdown (High readability) -->
                    <div class="h-2 bg-slate-950 rounded-full overflow-hidden w-full relative border border-slate-900/60 shadow-inner">
                      <div 
                        class="h-full rounded-full transition-all duration-500" 
                        style="width: {item.percentage}%; background-color: {item.color}; filter: brightness(1.05);"
                      ></div>
                    </div>
                  </div>
                {/each}
              {/if}
            </div>

          </div>

          <div class="border-t border-slate-900/60 pt-4 mt-6 flex items-center justify-between text-[10px] font-mono text-slate-550">
            <span>Listing {listData.length} unique elements</span>
            <span>GORM sqlite database</span>
          </div>

        </div>

      </div>
    {/if}

  </main>

  <!-- Footer -->
  <footer class="border-t border-slate-900/40 bg-slate-950 py-6 px-6 text-center text-[10px] text-slate-600 font-mono">
    IAM Ticket Dashboard // Backend SQLite 3 & Gin API // Developed Pair-wise
  </footer>

</div>

<!-- Detail Drawer Panel -->
{#if detailTicket}
  <!-- Overlay -->
  <div 
    onclick={() => detailTicket = null}
    class="fixed inset-0 bg-slate-950/70 backdrop-blur-sm z-45" 
    transition:fade
    role="button"
    tabindex="0"
  ></div>

  <!-- Drawer Body -->
  <div 
    class="fixed right-0 top-0 bottom-0 w-full max-w-lg bg-slate-950 border-l border-slate-900 z-50 p-6 flex flex-col justify-between shadow-2xl overflow-y-auto"
    transition:fly={{ x: 300, duration: 300 }}
  >
    <div class="space-y-6">
      
      <!-- Drawer Header -->
      <div class="flex items-center justify-between border-b border-slate-900/80 pb-4">
        <div>
          <span class="text-[10px] font-mono bg-indigo-500/10 border border-indigo-500/30 text-indigo-400 px-2 py-0.5 rounded font-bold">
            {detailTicket.ticket_id}
          </span>
          <h2 class="text-sm font-bold text-slate-200 mt-2">{detailTicket.subject}</h2>
        </div>
        <button 
          onclick={() => detailTicket = null}
          class="w-8 h-8 rounded-lg bg-slate-900 hover:bg-slate-800 text-slate-400 flex items-center justify-center border border-slate-800"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
            <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
          </svg>
        </button>
      </div>

      <!-- Drawer Content Grid -->
      <div class="space-y-4 text-xs">
        
        <!-- Section: Description -->
        <div class="space-y-1">
          <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Detailed Description</h4>
          <p class="bg-slate-900/40 border border-slate-900/60 rounded-xl p-3.5 text-slate-300 leading-relaxed font-light whitespace-pre-line font-mono select-text selection:bg-indigo-500/30">
            {detailTicket.DetailDescription || 'No description provided.'}
          </p>
        </div>

        <!-- Grid Attributes -->
        <div class="grid grid-cols-2 gap-4 pt-2">
          
          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Ticket Type</h4>
            <p class="text-slate-250 font-medium">{detailTicket.ticket_type}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Priority</h4>
            <p class="text-slate-250 font-medium capitalize">{detailTicket.priority}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Status</h4>
            <p class="text-slate-250 font-medium capitalize">{detailTicket.ticket_status}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Customer Name</h4>
            <p class="text-slate-250 font-medium">{detailTicket.customer_name || '-'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Assignee</h4>
            <p class="text-slate-250 font-medium">{detailTicket.Assignee || '-'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Assignment Group</h4>
            <p class="text-slate-250 font-medium">{detailTicket.GroupAssignee || '-'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Company / Country</h4>
            <p class="text-slate-250 font-medium">{detailTicket.company} ({detailTicket.country})</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Site / Region</h4>
            <p class="text-slate-250 font-medium">
              {detailTicket.site || '-'} / {detailTicket.region_site || '-'}
            </p>
          </div>

        </div>

        <!-- Categorization -->
        <div class="border-t border-slate-900/60 pt-4 grid grid-cols-3 gap-2">
          <div class="space-y-1">
            <h4 class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">Category T1</h4>
            <p 
              class="text-slate-300 font-medium text-[11px] truncate cursor-pointer"
              onmouseenter={(e) => { const t = detailTicket; if (t?.ProductT1) showTooltip(e, t.ProductT1); }}
              onmousemove={moveTooltip}
              onmouseleave={hideTooltip}
            >
              {detailTicket.ProductT1 || '-'}
            </p>
          </div>
          <div class="space-y-1">
            <h4 class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">Category T2</h4>
            <p 
              class="text-slate-300 font-medium text-[11px] truncate cursor-pointer"
              onmouseenter={(e) => { const t = detailTicket; if (t?.ProductT2) showTooltip(e, t.ProductT2); }}
              onmousemove={moveTooltip}
              onmouseleave={hideTooltip}
            >
              {detailTicket.ProductT2 || '-'}
            </p>
          </div>
          <div class="space-y-1">
            <h4 class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">Category T3</h4>
            <p 
              class="text-slate-300 font-medium text-[11px] truncate cursor-pointer"
              onmouseenter={(e) => { const t = detailTicket; if (t?.ProductT3) showTooltip(e, t.ProductT3); }}
              onmousemove={moveTooltip}
              onmouseleave={hideTooltip}
            >
              {detailTicket.ProductT3 || '-'}
            </p>
          </div>
        </div>

        <!-- Timestamps -->
        <div class="border-t border-slate-900/60 pt-4 space-y-2">
          <div class="flex justify-between items-center text-[11px]">
            <span class="text-slate-500">Created At</span>
            <span class="font-mono text-slate-300">{formatDate(detailTicket.CreatedDate)}</span>
          </div>
          {#if detailTicket.ResolvedAt}
            <div class="flex justify-between items-center text-[11px]">
              <span class="text-slate-500">Resolved At</span>
              <span class="font-mono text-slate-300">{formatDate(detailTicket.ResolvedAt)}</span>
            </div>
          {/if}
          {#if detailTicket.CompleteTime}
            <div class="flex justify-between items-center text-[11px]">
              <span class="text-slate-500">Completed At</span>
              <span class="font-mono text-slate-300">{formatDate(detailTicket.CompleteTime)}</span>
            </div>
          {/if}
          <div class="flex justify-between items-center text-[11px]">
            <span class="text-slate-500">Metadata Source</span>
            <span class="font-mono text-slate-400 bg-slate-900 border border-slate-800 px-2 py-0.5 rounded truncate max-w-xs">{detailTicket.SourceFile}</span>
          </div>
        </div>

      </div>

    </div>

    <div class="pt-6 border-t border-slate-900/80 mt-6 flex justify-end">
      <button 
        onclick={() => detailTicket = null}
        class="px-5 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-800 transition"
      >
        Close View
         </button>
    </div>
  </div>
{/if}

<!-- Floating Tooltip -->
{#if tooltipText}
  <div 
    class="fixed z-[9999] bg-slate-900/95 border border-slate-800 backdrop-blur-md text-slate-200 text-xs px-3 py-2 rounded-xl shadow-2xl max-w-xs pointer-events-none transition-all duration-75 break-words font-medium"
    style="left: {tooltipX}px; top: {tooltipY}px;"
  >
    {tooltipText}
  </div>
{/if}
