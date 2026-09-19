# PicoClaw y agent-harness

Zapia es un originador externo que llega por SSH a VM1 y ejecuta `agent-harness`. PicoClaw es el runtime: no es dueño del lifecycle ni de la evidencia.

Para coding gobernado, el `repository_root` debe ser un checkout Git con remoto `origin`. El harness crea un linked worktree en:

```text
<profile.workspace>/projects/<basename(repository_root)>
```

Ese worktree conserva el remoto, los objetos Git y el baseline. El harness registra lifecycle, policy, diff, tests, provenance y rollback; una respuesta textual de PicoClaw nunca sustituye la evidencia física.

No modificar upstream directamente. Las extensiones cm2labs deben mantenerse en esta capa extensible y comunicarse mediante adapters/contratos aprobados.
