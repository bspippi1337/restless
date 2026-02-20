(function(){
  const term = document.getElementById('term');
  const stamp = document.getElementById('buildstamp');
  if (stamp){
    const d = new Date();
    stamp.textContent = "Build: " + d.toISOString().slice(0,10);
  }
  if (!term) return;

  const lines = [
    "$ restless discover openai.com --save-profile openai",
    "✓ found: https://openai.com/.well-known/openapi.json (score 0.93)",
    "✓ found: https://openai.com/sitemap.xml (score 0.61)",
    "✓ extracted: 28 endpoints (verified: 19)",
    "✓ profile saved: profiles/openai.json",
    "",
    "$ restless console --profile openai",
    "restless> suggest",
    "• GET  /v1/models",
    "• POST /v1/responses   (stream)",
    "• POST /v1/embeddings",
    "",
    "restless> run GET /v1/models",
    "→ 200 OK   (12 ms)",
    "",
    "restless> save models",
    "✓ snippet saved: openai/models",
    "",
    "$ restless snippets list --profile openai",
    "• models     GET /v1/models",
    "• responses  POST /v1/responses (stream)"
  ];

  let i = 0, j = 0;
  const speed = 10;
  const lineDelay = 180;

  function tick(){
    if (i >= lines.length) return;
    const line = lines[i];
    if (j < line.length){
      term.textContent += line[j++];
      setTimeout(tick, speed);
      return;
    }
    term.textContent += "\n";
    i++; j = 0;
    setTimeout(tick, lineDelay);
  }
  setTimeout(tick, 350);
})();