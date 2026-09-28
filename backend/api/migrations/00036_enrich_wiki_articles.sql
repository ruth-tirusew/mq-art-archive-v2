-- +goose Up
-- The wiki articles seeded in 00013/00014 had teaser-length bodies (one sentence each) —
-- fine as list-page copy, but not a realistic reading experience on the article page
-- itself. This replaces each with a genuine multi-paragraph article on the same topic.

UPDATE articles SET
    body = 'This wiki is where Ethiopian artists write down what nobody teaches you in art school — the paperwork, the pricing, the shipping logistics, the contracts. Every entry here started as a submission from someone who actually lived through the problem it describes.

Anyone can read it. Verified artists and admins review new submissions before they go live, and an artist can revise an entry that''s out of date — prices change, offices move, rules get rewritten.

Start with the Legal and Materials sections if you''re new to selling work formally. If you have something to add, use the submission form linked from any article page, or draft a brand-new entry from your studio.

This is a living document, not a finished one. If something here is wrong or stale, say so — that''s the whole point of keeping it editable.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'welcome-to-mq';

UPDATE articles SET
    body = 'Under Ethiopian copyright law (Proclamation No. 410/2004), your copyright exists automatically the moment you finish a work — you don''t need to register anything for the right to exist. What registering with the EIPA gives you is evidence: a dated, official record you can point to if someone disputes who made a piece first.

Copyright covers the expression, not the idea. Painting your own version of a subject someone else has painted before — a market scene, a religious icon, a portrait style — is not infringement. Copying someone else''s specific composition, brushwork and color choices closely enough that it''s recognizably theirs, is.

Selling a painting doesn''t sell the copyright unless you say so in writing. By default you keep the right to reproduce it, print it, license the image commercially, and make derivative works — a buyer only owns the physical object.

If you''re licensing an image for a magazine, a book cover, or merchandise, always specify the term (one year vs. perpetual), the territory, and whether it''s exclusive. A one-page written agreement, even an email, is enough to make these terms enforceable later.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'copyright-basics';

UPDATE articles SET
    body = 'Fat over lean is the rule that keeps an oil painting from cracking: each successive layer should have more oil (and dry more slowly) than the one underneath. Painting a fast-drying, lean layer over a fatty one is the single most common cause of cracked or wrinkled paintings years later.

For underpainting, thin your paint with solvent alone (no oil) so it dries fast and matte. Build up medium (linseed or stand oil mixed with solvent) as you move into your mid-tones and final layers.

Addis Ababa''s dry season heat speeds drying dramatically compared to what most Western technique guides assume — a layer that takes three days to touch-dry in a humid climate can be dry in one here. Adjust your working schedule accordingly, and test a layer with your fingernail before painting over it.

Varnish only after the painting is fully cured — six months to a year for a normally-built painting, not just surface-dry. Varnishing too early traps solvent and moisture underneath and causes clouding.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'oil-painting-techniques';

UPDATE articles SET
    body = 'A consignment agreement means the gallery doesn''t own your work — they''re holding it to sell on your behalf, and you''re still legally the owner until it sells. That distinction matters if the gallery goes out of business: consigned work should be returned to you, not treated as gallery assets creditors can claim.

Commission splits in Addis galleries typically run 40/60 to 50/50 in the gallery''s favor for an unestablished artist, improving as your track record builds. Get the exact percentage, and whether it applies before or after framing/shipping costs, in writing.

Ask specifically about insurance: does the gallery insure consigned work while it''s in their space and in transit, and what happens if a piece is damaged or stolen. A gallery that can''t answer this clearly is a red flag.

Exclusivity clauses — agreeing not to sell through any other gallery, or not to sell directly to collectors, for the contract''s duration — are negotiable. Never sign an exclusivity window longer than 12 months without a clear exit clause.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'gallery-contracts';

UPDATE articles SET
    body = 'Piassa is still the first stop for most working painters — several long-standing shops near the Piassa roundabout stock imported oil paint, canvas by the meter, and basic hardware, though selection depends heavily on what''s recently cleared customs.

Mercato has cheaper raw materials if you''re willing to source and assemble yourself: raw linen and cotton duck by the meter from the textile section, and timber for stretcher bars from the wood merchants near the main gate. Expect to cut and join your own stretchers.

For pre-made stretched canvas and higher-end imported paint (Winsor & Newton, Old Holland), a couple of art-supply-focused shops around Bole carry stock but at a real markup versus Piassa — worth it mainly when you need something specific in a hurry.

Prices move with the exchange rate more than with local supply — imported pigment and canvas are effectively priced in USD even when quoted in birr. Buying in bulk when a shipment has just landed is the only real hedge most artists have found against this.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'addis-pigment-sources';

UPDATE articles SET
    body = 'Quote local clients in ETB and international clients in USD — don''t try to hold one exchange rate across a project that might take weeks, since birr depreciation during that time is a real cost you''ll otherwise eat yourself.

A simple base-rate approach: set an hourly or per-square-foot rate that covers materials plus your time at a wage you''d accept elsewhere, then apply a multiplier (many working Addis artists use 1.5–2x) for international commissions to account for shipping, payment processing fees, and the extra communication overhead of remote work.

The most common underpricing trap is anchoring to what local clients can pay and then applying that same number internationally out of habit, or discomfort asking for more. International collectors comparing you to artists in their own market are often expecting prices closer to that market''s, not Addis retail rates.

Get a deposit (30–50%) before starting any commission, local or international — non-refundable if the client cancels after you''ve begun. This is standard practice, not an insult to ask for.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'pricing-local-vs-international';

UPDATE articles SET
    body = 'For a single painting or a small batch, DHL or FedEx door-to-door is usually simpler than freight forwarding — freight forwarders make more sense above roughly ten paintings or for anything large or heavy (sculpture, big panels) where per-kilogram courier rates get punishing.

Customs paperwork needs a commercial invoice describing the work, materials, dimensions and declared value — undervaluing to reduce duties is common practice but puts you at risk if a piece is lost or damaged, since insurance claims are capped at the declared value.

Crate or hard-case anything valuable rather than relying on courier packaging alone. A basic plywood travel frame around a stretched canvas, with corner protection, survives handling far better and costs a fraction of what a damaged, uninsurable painting costs you.

Collectors abroad are usually the ones asking about import duty on their end — have an answer ready. Most countries treat original artwork favorably for duty purposes, but courier customs agents don''t always know this and may miscode the shipment as general merchandise unless the paperwork is explicit.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'shipping-works-abroad';

UPDATE articles SET
    body = 'Ethiopian copyright law has no codified "fair use" doctrine in the way U.S. law does — there''s no fixed four-factor test. Instead the law permits specific, narrower exceptions: quotation, teaching, and personal use, each defined more narrowly than most artists assume.

Using a reference photograph you didn''t take — a market scene, a portrait, a news photo — as the visual basis for a painting sits in a genuinely unclear area of the law. The safer practice used by most working artists here is to substantially transform the source (different medium, composition changes, your own color and mark-making) rather than rely on any fair-use-style defense.

Sampling or referencing another living artist''s distinctive style is not itself infringement — style isn''t protected, only specific expression is. But recreating a specific painting''s composition closely, even in a different medium, likely is.

When in doubt, ask. Most Ethiopian photographers and artists will grant a reference-use permission for free or a small fee if you simply ask and credit them — it''s a much smaller conversation than the legal risk of not asking.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'fair-use-amharic';

UPDATE articles SET
    body = 'The Ethiopian Intellectual Property Authority handles both copyright deposit and trademark registration — for an individual artist, it''s the copyright deposit process that matters, establishing an official dated record of a specific work.

You''ll need: a completed application form (available at the EIPA office or, increasingly, online), a physical or high-resolution digital copy of the work, proof of identity, and the filing fee, which is modest relative to the value it provides as evidence.

Registration is not required for copyright to exist — it exists automatically on creation — but it matters practically if you ever need to prove authorship and a date in a dispute, or if a gallery or licensee asks for registration as part of doing business with you.

Processing takes a few weeks in practice, longer if the application is incomplete. Bring more documentation than the checklist strictly requires — a rejected first submission adds real delay, since EIPA doesn''t process incrementally.',
    reading_time = 2,
    updated_at = NOW()
WHERE slug = 'eipa-registration';

-- +goose Down
-- Not reversible to the exact prior one-sentence bodies (not worth preserving them);
-- down leaves the enriched content in place.
SELECT 1;
