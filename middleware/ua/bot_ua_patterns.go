package ua

// DefaultBotPatterns is a categorized map of known bot, scanner, scraper, and
// crawler User-Agent substrings. Users pick categories via "defaults": ["scanners", "scrapers", ...]
// This is partially AI generated. Well, most of it is AI generated
var DefaultBotPatterns = map[string][]string{
	// HTTP client libraries — programmatic access, not browsers
	"http-clients": {
		// Python
		"python-requests", "python-urllib", "python-httpx", "aiohttp",
		"httpx", "urllib3", "pycurl",
		// Go
		"go-http-client", "fasthttp",
		// Java
		"java/", "apache-httpclient", "okhttp", "jersey/",
		// Ruby
		"ruby", "faraday", "typhoeus", "httparty", "rest-client",
		// PHP
		"guzzlehttp", "php/", "symfony",
		// .NET
		"httpclient", "restsharp",
		// Rust
		"reqwest", "hyper",
		// Node.js
		"node-fetch", "axios", "undici", "got/",
		// Perl
		"libwww-perl", "lwp-request", "www-mechanize",
		// Generic
		"http_request2", "winhttp",
	},

	// CLI tools — manual or scripted
	"cli-tools": {
		"curl", "wget", "httpie", "aria2",
		"fetch/", "lynx", "links", "elinks", "w3m",
	},

	// Vulnerability scanners and pentesting tools
	"scanners": {
		// General purpose
		"nikto", "nuclei", "zap", "burpsuite", "burp",
		"acunetix", "nessus", "openvas", "qualys",
		"webinspect", "appscan", "arachni",
		"skipfish", "w3af", "paros", "webshag",
		"vega", "grabber", "wapiti", "ratproxy",
		"jaeles", "caido",
		// Brute force / fuzzing
		"dirbuster", "gobuster", "ffuf", "feroxbuster",
		"wfuzz", "dirb", "patator",
		"hydra", "medusa", "hashcat",
		// SQL injection
		"sqlmap", "sqlninja", "havij",
		// XSS
		"xsstrike", "dalfox", "xsser",
		// Command injection
		"commix",
		// CMS scanners
		"wpscan", "joomscan", "droopescan", "cmsmap",
		// Network
		"nmap", "masscan", "zmap", "unicornscan",
	},

	// Internet-wide scanners and research bots
	"recon": {
		"zgrab", "censys", "shodan",
		"binaryedge", "leakix",
		"onyphe", "netcraft",
		"internetmeasur", "internet-measurement",
		"rapid7", "shadowserver",
		"project25499",
		"expanseinc", "paloaltonetworks",
	},

	// Scraping frameworks and automation
	"scrapers": {
		// Frameworks
		"scrapy", "colly", "mechanize", "twill",
		"goutte", "cheerio",
		"httrack", "teleport", "webcopier",
		"sitesucker", "pavuk", "offline explorer",
		// Commercial
		"crawlera", "zyte",
		"apify", "scrapingbee", "scrapingant",
		"webscrapingapi", "scraperapi",
		"brightdata", "luminati",
		// Data extraction
		"import.io", "diffbot", "parsehub", "octoparse",
		"contentking",
	},

	// Headless browsers and automation tools
	"headless": {
		"headlesschrome", "phantomjs",
		"selenium", "puppeteer", "playwright",
		"slimerjs", "casperjs", "splash",
		"nightmare", "cypress",
		"webdriver", "chromedriver", "geckodriver",
	},

	// SEO / marketing crawlers (some legitimate but noisy)
	"seo-bots": {
		"ahrefsbot", "semrushbot", "dotbot",
		"mj12bot", "rogerbot", "blexbot",
		"serpstatbot", "megaindex",
		"seokicks", "sistrix", "linkdexbot",
		"screaming frog", "seobility",
		"seostar", "df bot",
		"backlinkcrawler", "domainreanimator",
		"linkpadbot", "seoscanners",
	},

	// Known malicious or suspicious UAs
	"malicious": {
		"zmeu", "morfeus",
		"exploit", "payload", "shellshock",
		"muhstik", "mirai",
		"hajime", "tsunami", "xmrig",
		"coinminer", "cryptominer",
		"dirtjumper", "blackenergy",
	},

	// Generic bot indicators — broad substring matches
	"generic": {
		"bot/", "spider/", "crawl/", "crawler",
		"scan/", "scanner",
		"archiver", "collector", "fetcher",
		"monitor", "checker", "validator",
		"slurp", "spyder", "harvest",
		"extractor", "sucker", "stripper",
		"grab", "leech", "siphon",
	},

	// Social media / preview bots (usually legitimate — opt-in)
	"social": {
		"facebookexternalhit", "facebot",
		"twitterbot", "linkedinbot",
		"slackbot", "telegrambot",
		"discordbot", "whatsapp",
		"pinterestbot", "applebot",
		"skypeuripreview",
	},

	// Search engine bots (usually legitimate — opt-in)
	"search-engines": {
		"googlebot", "bingbot", "yandexbot",
		"baiduspider", "duckduckbot",
		"sogou", "exabot", "ia_archiver",
		"archive.org_bot",
	},

	// AI training data crawlers
	"ai-crawlers": {
		"gptbot", "chatgpt-user", "oai-searchbot",
		"ccbot", "anthropic-ai", "claude-web",
		"cohere-ai", "bytespider", "petalbot",
		"amazonbot", "meta-externalagent",
		"img2dataset", "commoncrawl",
	},
}
