# Contexto operativo: PicoClaw y Hermes

Última actualización: 2026-09-05
Responsable: Claudio Maza / cm2labs

Este documento permite retomar el trabajo sin depender del historial de chat.

## 1. PicoClaw

### Repositorios

- Desarrollo privado: `claudiomaza/picoclaw-031-vF-prod`
- Upstream original: `sipeed/picoclaw`
- Backup de configuración/runtime: `claudiomaza/picoclaw-backup-config`
- Checkout local VM1: `/home/ubuntu/picoclaw-source-diagnose`
- Runtime operativo VM1: `/home/ubuntu/picoclaw`

Remotes esperados en el checkout:

```text
origin   -> https://github.com/sipeed/picoclaw
personal -> https://github.com/claudiomaza/picoclaw-031-vF-prod.git
```

La rama de trabajo y publicación es `main`. `origin/main` es solo upstream: nunca hacer pull ciego.

### Cómo retomar PicoClaw

```bash
cd /home/ubuntu/picoclaw-source-diagnose
git status --short --branch
git fetch origin
git fetch personal
git log --oneline personal/main..origin/main
git diff personal/main...origin/main
```

Interpretación:

- `personal/main`: nuestra implementación estable.
- `origin/main`: cambios nuevos del proyecto original.
- Los commits upstream se revisan y contrastan antes de incorporarlos.
- Las actualizaciones se integran selectivamente, preservando los cambios de cm2labs.

Antes de modificar código:

1. Fijar esta carpeta como proyecto.
2. Confirmar que existe `.git`.
3. Ejecutar `git status --short`.
4. Registrar mentalmente o documentar el HEAD inicial.
5. Trabajar con diff verificable.

Después de modificar:

```bash
git diff --check
git status --short
git diff --stat
go test ./pkg/tools ./pkg/agent
```

Solo reportar cambios que aparezcan en `git diff` o `git status`.

### Cambios propios relevantes

La línea privada parte del upstream en:

```text
bbf6893ca7 feat(models): add configurable default fallback chain (#3200)
```

Commits cm2labs relevantes:

- `c278f60545` — baseline del proyecto original.
- `aa36319745` — skills de PicoClaw.
- `fe94394eb7` — sincronización de `backup-picoclaw`.
- `35a911ffce`, `5f2d2b0dec`, `f3e554223c` — actualizaciones de GitHub Actions.
- `ad909db449`, `d92864db28`, `238654da56`, `ebdddcae2d` — dependencias Go.
- `c1dfe19b55` — routing de `spawn` mediante el manager durable.

El diseño de subagentes durables está en `BLUEPRINT.md`.

### PicoClaw operativo

Agentes configurados:

```text
planner, orchestrator, tester, reviewer, coder
```

Servicios:

```text
picoclaw.service
proxy-a2a.service
proxy-a2a.service
```

Configuración A2A:

```text
PICOCLAW_CONFIG=/home/ubuntu/picoclaw/config/config.json
PICOCLAW_A2A_ADDR=127.0.0.1:8644
```

El backup runtime usa un overlay para no duplicar el código tracked:
`claudiomaza/picoclaw-backup-config`.

## 2. Hermes

### Repositorios y runtime

- Fork público: `claudiomaza/hermes-agent`
- Upstream: `NousResearch/hermes-agent`
- Espejo privado: `claudiomaza/hermes-agent-private`
- Runtime Hermes VM2: `/home/ubuntu/.hermes`
- Checkout fuente VM2: `/home/ubuntu/.hermes/hermes-agent`
- Rama local actual: `cm2labs/hermes-roundrobin`
- Backup portable: `claudiomaza/hermes-recovery-backup`

El fork público se conserva para seguir el upstream. El espejo privado es independiente y no se sincroniza solo.

### Seguimiento de cambios Hermes

Antes de trabajar:

```bash
cd /home/ubuntu/.hermes/hermes-agent
git status --short --branch
git log --oneline -10
```

Para revisar upstream:

```bash
git fetch origin
git log --oneline HEAD..origin/main
git diff HEAD...origin/main
```

Hermes dispone de:

- `hermes project init` para crear un proyecto Git con baseline.
- `hermes project verify` para revisar estado, diff y evidencia.
- soporte de worktrees.
- detección de project root Git.
- verificación de mutaciones de archivos al finalizar turnos.

Implementado en esta línea: PicoClaw ejecuta un preflight Git por turno y Hermes ejecuta un preflight para workspaces de código. Ambos inicializan Git de forma controlada cuando falta, excluyen secretos/estado runtime del baseline y registran el estado inicial/final. La verificación debe mantenerse cubierta por tests.

## 3. Monitoreo y backups

Estado actual:

- PR abiertos de los repositorios actuales: `0`.
- PicoClaw privado tiene backup overlay verificado.
- Hermes tiene backup portable verificado con secretos autorizado por el usuario.
- Dropbox no forma parte del flujo operativo.
- No existe todavía un monitor automático confirmado para detectar nuevos commits upstream.

Revisión manual recomendada:

1. `git fetch origin` en PicoClaw y Hermes.
2. Comparar commits y diff contra la rama privada/local.
3. Revisar cambios upstream antes de integrar.
4. Ejecutar tests.
5. Crear backup después de cambios estructurales.
6. Actualizar este documento y `BLUEPRINT.md` si cambia el diseño.

## 4. Cómo invocar este contexto

En una nueva sesión pedir, por ejemplo:

> Retomemos el proyecto PicoClaw/Hermes. Leé `PROJECT-CONTEXT-PICOCLAW-HERMES.md` y `BLUEPRINT.md`, revisá primero el estado Git local, compará upstream contra `personal/main`, y no modifiques nada hasta mostrarme el diff y el plan.

Para retomar solo PicoClaw:

> Retomemos PicoClaw desde `PROJECT-CONTEXT-PICOCLAW-HERMES.md`. Verificá `personal/main` contra `origin/main`, revisá cambios upstream y continuá desde el último commit privado.

Para retomar solo Hermes:

> Retomemos Hermes desde `PROJECT-CONTEXT-PICOCLAW-HERMES.md`. Revisá VM2, la rama `cm2labs/hermes-roundrobin`, el estado Git y el backup portable antes de trabajar.

Para una revisión sin cambios:

> Hacé un monitoreo de upstream de PicoClaw y Hermes: fetch, commits nuevos, diff, riesgos, tests recomendados y reporte. No hagas merge ni push.

Para integrar cambios aprobados:

> Aplicá selectivamente los cambios upstream de PicoClaw/Hermes según el contexto documentado, ejecutá tests, actualizá la documentación y pedime confirmación antes del push.

## 5. Regla de continuidad

Nunca asumir que el checkout local está actualizado. Siempre comenzar por:

```text
proyecto -> carpeta -> Git -> HEAD -> status -> upstream -> diff -> plan -> cambios -> tests -> commit -> backup
```
