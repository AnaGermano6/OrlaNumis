<template>
  <section class="card min-w-0 overflow-hidden">
    <h2 class="mb-2">Guia de iniciação ao colecionismo de moedas</h2>
    <p class="mb-6 text-body text-text-secondary">
      Tudo o que precisa de saber para começar a colecionar moedas com confiança.
    </p>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim" open>
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Primeiros passos na aplicação</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>Se esta for uma conta nova, pode abrir este guia a partir do aviso do primeiro início de sessão ou, a qualquer momento, na barra lateral em <strong>Primeiros passos</strong>.</p>

        <h4>Primeiros passos recomendados</h4>
        <ol>
          <li>Crie a sua conta e abra <strong>Definições → Cópias de segurança</strong>.</li>
          <li>Transfira o <strong>Modelo CSV</strong>.</li>
          <li>Preencha as suas moedas numa folha de cálculo (Excel, Google Sheets, Numbers) e exporte como CSV.</li>
          <li>Importe o CSV em <strong>Definições → Cópias de segurança → Importar</strong>.</li>
          <li>Reveja as moedas importadas e carregue as fotografias na página de detalhe de cada moeda.</li>
        </ol>

        <h4>Colunas do CSV</h4>
        <p>Só a coluna <code>name</code> é obrigatória. Todos os outros campos são opcionais.</p>
        <table>
          <thead>
            <tr><th>Coluna</th><th>Tipo</th><th>Notas</th></tr>
          </thead>
          <tbody>
            <tr><td><code>name</code></td><td>texto</td><td>Obrigatória</td></tr>
            <tr><td><code>category</code></td><td>texto</td><td>Uma das categorias configuradas pelo administrador (predefinidas: Roman, Greek, Byzantine, Modern, Other)</td></tr>
            <tr><td><code>material</code></td><td>texto</td><td>Gold, Silver, Bronze, Copper, Electrum, Other</td></tr>
            <tr><td><code>denomination</code>, <code>ruler</code>, <code>mint</code>, <code>grade</code></td><td>texto</td><td>Campos de texto livre</td></tr>
            <tr><td><code>era</code></td><td>texto</td><td>Uma das épocas configuradas pelo administrador (predefinidas: <code>ancient</code>, <code>medieval</code>, <code>modern</code>). Indique as datas reais da moeda em <code>ruler</code> ou <code>notes</code> — uma época fora da lista configurada não corresponde aos filtros de época.</td></tr>
            <tr><td><code>weightGrams</code>, <code>diameterMm</code>, <code>purchasePrice</code>, <code>currentValue</code>, <code>soldPrice</code></td><td>número</td><td>Use valores decimais. Se ficar em branco, <code>currentValue</code> assume o valor de <code>purchasePrice</code>.</td></tr>
            <tr><td><code>purchaseDate</code>, <code>soldDate</code></td><td>data</td><td>Use <code>AAAA-MM-DD</code></td></tr>
            <tr><td><code>isWishlist</code>, <code>isSold</code>, <code>isPrivate</code></td><td>booleano</td><td>Use true/false</td></tr>
            <tr><td><code>purchaseLocation</code>, <code>soldTo</code>, <code>vendorSku</code>, <code>vendorInvoice</code></td><td>texto</td><td>Registo da aquisição e da venda</td></tr>
            <tr><td><code>obverseInscription</code>, <code>reverseInscription</code>, <code>obverseDescription</code>, <code>reverseDescription</code>, <code>rarityRating</code></td><td>texto</td><td>Dados de atribuição, se já os tiver</td></tr>
            <tr><td><code>notes</code>, <code>referenceUrl</code>, <code>referenceText</code></td><td>texto</td><td>Metadados opcionais (campos antigos de referência)</td></tr>
          </tbody>
        </table>

        <h4>Nomes alternativos das colunas</h4>
        <p>Se a sua folha de cálculo já usa outros cabeçalhos, a importação aceita-os sem ter de os mudar:</p>
        <table>
          <thead>
            <tr><th>Coluna</th><th>Também aceite</th></tr>
          </thead>
          <tbody>
            <tr><td><code>name</code></td><td><code>coinName</code>, <code>title</code></td></tr>
            <tr><td><code>ruler</code></td><td><code>emperor</code>, <code>issuer</code></td></tr>
            <tr><td><code>era</code></td><td><code>date</code></td></tr>
            <tr><td><code>weightGrams</code>, <code>diameterMm</code></td><td><code>weight</code>/<code>weight_g</code>, <code>diameter</code>/<code>diameter_mm</code></td></tr>
            <tr><td><code>grade</code></td><td><code>condition</code></td></tr>
            <tr><td><code>purchasePrice</code>, <code>currentValue</code></td><td><code>pricePaid</code>, <code>estimatedValue</code></td></tr>
            <tr><td><code>purchaseDate</code></td><td><code>acquiredDate</code></td></tr>
            <tr><td><code>purchaseLocation</code></td><td><code>store</code>, <code>dealer</code></td></tr>
            <tr><td><code>vendorSku</code>, <code>vendorInvoice</code></td><td><code>sku</code>/<code>vendorCode</code>, <code>invoice</code>/<code>invoiceNumber</code></td></tr>
            <tr><td><code>rarityRating</code></td><td><code>reference</code>, <code>catalogReference</code></td></tr>
            <tr><td><code>referenceUrl</code></td><td><code>url</code></td></tr>
            <tr><td><code>isWishlist</code>, <code>isSold</code>, <code>isPrivate</code></td><td><code>wishlist</code>, <code>sold</code>, <code>private</code></td></tr>
          </tbody>
        </table>
        <p>As linhas sem <code>name</code> são ignoradas e a importação indica quantas foram ignoradas.</p>

        <h4>Exemplo de CSV</h4>
        <pre>name,category,material,denomination,ruler,era,mint,weightGrams,diameterMm,grade,purchasePrice,currentValue,purchaseDate,purchaseLocation,isWishlist
Augustus Denarius,Roman,Silver,Denarius,Augustus (27 BC - 14 AD),ancient,Rome,3.82,19.5,VF,450,600,2024-03-15,Heritage Auctions,false
Constantius II Follis,Roman,Bronze,Follis,Constantius II (337-361 AD),ancient,Antioch,2.90,18.1,F,35,45,2025-01-20,Local Show,false</pre>

        <h4>Onde estão as coisas nas Definições</h4>
        <table>
          <thead>
            <tr><th>Separador</th><th>O que contém</th></tr>
          </thead>
          <tbody>
            <tr><td><strong>Conta</strong></td><td>Perfil, palavra-passe, Pushover, Moeda do dia, fornecedores de início de sessão associados, início de sessão biométrico</td></tr>
            <tr><td><strong>Ligações</strong></td><td>Credenciais do NumisBids, dos CNG Auctions e do ParcelApp</td></tr>
            <tr><td><strong>Aparência</strong></td><td>Tema, idioma e preferências de apresentação</td></tr>
            <tr><td><strong>Dados</strong></td><td>Etiquetas e conjuntos abertos, locais de armazenamento, localizações personalizadas de casas da moeda, migração de referências de catálogo</td></tr>
            <tr><td><strong>Chaves</strong></td><td>Chaves de API para clientes de IA externos</td></tr>
            <tr><td><strong>Ferramentas</strong></td><td>Definições do processador de imagens e utilizadores bloqueados</td></tr>
            <tr><td><strong>Conversas</strong></td><td>Conversas guardadas com o agente</td></tr>
            <tr><td><strong>Envios</strong></td><td>Envios ainda não entregues</td></tr>
            <tr><td><strong>Cópias de segurança</strong></td><td>Importação da coleção, exportação ZIP, exportação do catálogo em PDF, modelo CSV</td></tr>
            <tr><td><strong>Ajuda</strong></td><td>Este guia</td></tr>
          </tbody>
        </table>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Usar as funcionalidades do Aurearia</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>O Aurearia permite catalogar a coleção, planear aquisições, acompanhar leilões, partilhar com outros colecionadores e pesquisar com ajuda de IA. Estes fluxos são opcionais: comece por importar a coleção ou pela Captura rápida e ative as integrações quando precisar.</p>

        <h4>Registo rápido</h4>
        <ul>
          <li><strong>Por onde começar</strong> — Os dois fluxos de registo estão em <strong>Identificar moeda</strong> na barra lateral; a lista de rascunhos é a ação <strong>Todos os rascunhos</strong> nessa página.</li>
          <li><strong>Identificação de moedas</strong> — Fotografe uma moeda ou uma cápsula NGC Ancients, reveja os dados extraídos e guarde o resultado na coleção ou na lista de desejos. <strong>Identificar moeda → Histórico de análises</strong> guarda as identificações aprofundadas anteriores, incluindo os resultados e as consultas a cada fornecedor.</li>
          <li><strong>Notas por voz</strong> — Nos navegadores compatíveis, use <strong>Ditar notas</strong> na identificação rápida ou na Análise aprofundada de uma moeda guardada, reveja a transcrição (no idioma da interface) e submeta com a ação habitual. O fornecedor do navegador ou do dispositivo pode processar o áudio; o Aurearia não o guarda. Pode sempre escrever.</li>
          <li><strong>Captura rápida</strong> — Guarde fotografias e notas tiradas numa feira como rascunhos, retome-os mais tarde e promova um rascunho completo para a coleção ou para a lista de desejos. Os rascunhos ficam fora dos totais da coleção, da lista de desejos, das estatísticas e da saúde até serem promovidos.</li>
        </ul>

        <h4>Leilões</h4>
        <ul>
          <li><strong>Importação manual</strong> — Cole o URL de um lote do NumisBids ou dos CNG Auctions na página Leilões.</li>
          <li><strong>Sincronização das listas de observação</strong> — Adicione as credenciais do NumisBids e/ou do CNG em <strong>Definições → Ligações</strong> e use <strong>Sincronizar listas de observação</strong> na página Leilões.</li>
          <li><strong>Capacidades de cada fornecedor</strong> — Os CNG Auctions permitem uma sincronização mais completa dos leilões que alojam e podem detetar automaticamente se ganhou ou perdeu, quando o CNG fornece esses dados. O NumisBids, por agora, só permite acompanhar a lista de observação e as importações; depois do leilão, verifique no NumisBids e atualize manualmente o resultado, a licitação vencedora e a licitação máxima.</li>
          <li><strong>Filtros por origem</strong> — Filtre os lotes de leilão por Todos, NumisBids ou CNG, além do estado.</li>
          <li><strong>Segurança</strong> — As palavras-passe dos fornecedores de leilões são guardadas encriptadas no servidor. Os valores antigos em texto simples são migrados automaticamente da próxima vez que forem guardados ou usados numa sincronização.</li>
        </ul>

        <h4>Partilha e descoberta</h4>
        <ul>
          <li><strong>Vitrinas públicas</strong> — Crie seleções públicas com ligações partilháveis em <strong>Vitrinas</strong> na barra lateral. Só são expostos campos seguros para o público; as notas privadas, os preços e a análise por IA ficam ocultos.</li>
          <li><strong>Seguidores</strong> — Siga outros colecionadores em <strong>Seguidores</strong>, veja as galerias que partilham e comente ou classifique moedas. A sua própria visibilidade é controlada por moeda e pela definição de privacidade do perfil em Definições → Conta.</li>
          <li><strong>Alertas de pesquisa da lista de desejos</strong> — Guarde critérios de pesquisa, execute alertas de forma agendada ou manual, reveja os candidatos, rejeite as más correspondências e converta os candidatos aceites em itens da lista de desejos.</li>
          <li><strong>Saúde da coleção</strong> — Use as vistas de saúde nas Estatísticas para acompanhar a evolução da pontuação, imagens em falta, lacunas de metadados, lacunas de análise por IA e melhorias sugeridas.</li>
          <li><strong>Tabuleiro de museu</strong> — Use a vista Tabuleiro para dispor as moedas visualmente, com espaçamento proporcional ao tamanho, como num gabinete. As moedas que possui mostram a data de compra em <code>AAAA-MM-DD</code>; os itens da lista de desejos aparecem mais transparentes e mostram <code>TBD</code>.</li>
          <li><strong>Mapa das casas da moeda</strong> — Use Estatísticas → Mapa para ver as localizações das casas da moeda e a distribuição geográfica das moedas com casa da moeda associada.</li>
          <li><strong>Moeda do dia</strong> — Ative as notificações diárias da moeda em destaque em Definições → Conta para redescobrir moedas da sua coleção.</li>
          <li><strong>Fornecedores de início de sessão associados</strong> — Depois de iniciar sessão localmente, abra Definições → Conta → Fornecedores de início de sessão associados para associar o Microsoft Entra ID, o Pocket ID ou outro fornecedor OIDC configurado.</li>
          <li><strong>Início de sessão biométrico</strong> — Registe uma chave de acesso em Definições → Conta → Início de sessão biométrico para entrar com Face ID, Touch ID, Windows Hello ou uma chave de segurança.</li>
        </ul>

        <h4 data-testid="help-coin-copilot">Coin Copilot (beta)</h4>
        <ul>
          <li><strong>O que é</strong> — Quando o administrador ativa o Coin Copilot, o mesmo painel de conversa planeia vários passos, usa ferramentas só de leitura (pesquisas na coleção, em comerciantes, em leilões e de tendências de preços) e depois responde. Nunca altera as suas moedas, a lista de desejos nem compras.</li>
          <li><strong>Como o reconhecer</strong> — A conversa mostra a etiqueta <strong>Beta · Coin Copilot</strong>, uma lista de passos enquanto trabalha e um cartão de evidências das fontes para cada pesquisa. Sem essa etiqueta está a usar o Agente de moedas original, descrito abaixo.</li>
          <li><strong>Perguntas e resultados</strong> — Pode fazer uma pausa para pedir esclarecimentos quando o pedido é ambíguo. As pesquisas de mercado devolvem 5 anúncios por predefinição; peça mais para obter até 10.</li>
          <li><strong>Limitações</strong> — A pesquisa de feiras de moedas ainda não está disponível no Copilot. A Análise aprofundada a partir da conversa só funciona se o administrador o permitir.</li>
          <li><strong>Saber mais</strong> — Consulte o guia do colecionador <em>How Coin Copilot Works</em> na documentação do projeto.</li>
        </ul>

        <h4>Agente de moedas e notas</h4>
        <ul>
          <li><strong>Agente de moedas</strong> — Quando o Coin Copilot está desligado, abra a conversa em <strong>Agente</strong> na barra lateral (ou no botão arrastável no modo PWA) para pesquisar anúncios, encontrar feiras, fazer perguntas sobre o portefólio ou pedir ajuda a arrumar a coleção, por exemplo que moedas não têm época.</li>
          <li><strong>Sugestões iniciais</strong> — O painel inclui exemplos de aquisição e uma sugestão de arrumação para moedas sem época.</li>
          <li><strong>Guardar nas notas</strong> — As respostas terminadas do assistente podem ser revistas em markdown, editadas, pré-visualizadas e guardadas nas Notas só depois de clicar em Criar nota.</li>
        </ul>

        <h4>Conjuntos e propostas do agente</h4>
        <ul>
          <li><strong>Tipos de conjunto</strong> — Crie conjuntos normais para organização manual, conjuntos de objetivo em que os itens da lista de desejos representam o que falta, conjuntos inteligentes a partir de regras e conjuntos do agente a partir de uma proposta revista.</li>
          <li><strong>Fluxo de propostas do agente</strong> — Criar um conjunto do agente envia um pedido em vez de criar logo o conjunto. Quando a proposta estiver pronta, uma notificação abre a página de revisão.</li>
          <li><strong>Revisão humana</strong> — Reveja o âmbito, as posições e as notas do agente; edite, gere novamente com comentários, rejeite ou aprove. Só a aprovação cria o conjunto.</li>
          <li><strong>Correspondência automática</strong> — Os conjuntos do agente aprovados fazem corresponder automaticamente as moedas da coleção à lista proposta, tal como o registo de imperadores romanos.</li>
        </ul>

        <h4>Vistas de estatísticas</h4>
        <ul>
          <li><strong>Resumo</strong> — Totais da coleção, classificações, gráficos de distribuição e visão geral do portefólio.</li>
          <li><strong>Cronologia e Mapa</strong> — Histórico cronológico da coleção e geografia das casas da moeda das moedas mapeadas.</li>
          <li><strong>Saúde</strong> — Indicadores de imagens em falta, lacunas de metadados, lacunas de análise por IA, avaliações desatualizadas e sugestões de arrumação.</li>
          <li><strong>Detalhes de valor e Repartição do investimento</strong> — Evolução do valor, distribuição, desempenho por ano de aquisição, maiores subidas/descidas e avaliações desatualizadas.</li>
          <li><strong>Máquina do tempo</strong> — Recue no tempo para ver a coleção tal como estava em qualquer data.</li>
          <li><strong>Registo de imperadores</strong> — Progresso opcional por imperador/figura romana, com ações de pesquisa do agente para alvos em falta ou sugeridos. Está em <strong>Conjuntos → Imperadores</strong>, não nas Estatísticas, e aparece depois de ativado para a sua conta.</li>
        </ul>

        <h4>Assistentes de IA</h4>
        <ul>
          <li><strong>Análise da moeda</strong> — Execute a análise por modelo de visão sobre as fotografias do anverso/reverso na página de detalhe, para ajudar na atribuição e na descrição.</li>
          <li><strong>Identificação aprofundada</strong> — Identificação mais demorada com vários fornecedores; os resultados e as consultas de cada tarefa ficam em <strong>Identificar moeda → Histórico de análises</strong>.</li>
          <li><strong>Assistente de classificação</strong> — Estime o grau de conservação a partir de fotografias. Trate-o como uma segunda opinião, não como substituto de uma classificação profissional.</li>
          <li><strong>Tendências de preços</strong> — Análise da tendência de mercado de uma moeda ou tipo.</li>
          <li><strong>Análise de lacunas</strong> — Sugestões do que falta na coleção, com base no que já tem.</li>
          <li><strong>Guia de fotografia</strong> — Comentários sobre a qualidade das imagens, para que a nova fotografia corrija o que interessa.</li>
          <li><strong>Lotes semelhantes</strong> — Encontre lotes de leilão que correspondam a uma moeda ou a um item da lista de desejos.</li>
        </ul>

        <h4>Organizar e vender</h4>
        <ul>
          <li><strong>Etiquetas personalizadas</strong> — Crie etiquetas com cores em <strong>Definições → Dados</strong>, em Etiquetas e conjuntos abertos, e filtre a coleção por elas.</li>
          <li><strong>Operações em lote</strong> — Entre no modo de seleção na barra da coleção para etiquetar, adicionar a conjuntos ou alterar várias moedas de uma vez, sem abrir cada uma.</li>
          <li><strong>Locais de armazenamento</strong> — Registe onde cada moeda está fisicamente; geridos em <strong>Definições → Dados</strong>.</li>
          <li><strong>Diário e cronologia</strong> — Cada moeda tem um diário datado e uma cronologia do que lhe aconteceu, incluindo alterações feitas por ferramentas externas.</li>
          <li><strong>Notas</strong> — As notas de pesquisa independentes estão em <strong>Notas</strong>; as respostas do agente podem ser guardadas lá.</li>
          <li><strong>Moedas vendidas</strong> — Venda uma moeda na página de detalhe para registar o preço e o comprador; <strong>Vendidas</strong> acompanha depois o lucro ou prejuízo face ao que pagou.</li>
        </ul>

        <h4>Envios e lembretes</h4>
        <ul>
          <li><strong>Seguimento de envios</strong> — Acompanhe uma compra a caminho na página <strong>Seguimento de envio</strong> da moeda. Se adicionar uma chave de API do ParcelApp em <strong>Definições → Ligações</strong>, o seguimento pode ser criado a partir de um número de seguimento e sincronizado automaticamente quando o administrador o ativa.</li>
          <li><strong>Resumo de envios</strong> — <strong>Definições → Envios</strong> lista todos os envios ainda não entregues.</li>
          <li><strong>Lembretes de compra</strong> — Agende um lembrete para uma moeda que pretende comprar e receba uma notificação quando chegar a data.</li>
          <li><strong>Calendário</strong> — O <strong>Calendário</strong> mostra as datas dos leilões dos lotes acompanhados, juntamente com os seus próprios eventos, como feiras e tarefas de pesquisa.</li>
        </ul>

        <h4>Cópias de segurança e exportação</h4>
        <ul>
          <li><strong>Exportar ZIP</strong> — <strong>Definições → Cópias de segurança</strong> exporta a coleção completa, incluindo imagens, num arquivo que pode ser restaurado.</li>
          <li><strong>Exportar PDF</strong> — Gere, no mesmo separador, um catálogo para seguro e proveniência com fotografias e dados estruturados.</li>
          <li><strong>Importar</strong> — O mesmo separador importa uma cópia de segurança JSON ou um CSV e disponibiliza o modelo CSV descrito acima em Primeiros passos.</li>
          <li><strong>Conversas guardadas</strong> — <strong>Definições → Conversas</strong> guarda as conversas anteriores com o agente para as reabrir ou eliminar.</li>
        </ul>

        <h4>Captura, imagens e utilização offline</h4>
        <ul>
          <li><strong>Captura com a câmara</strong> — A Captura rápida, a identificação de moedas e os carregamentos de imagens só ligam a câmara depois de tocar numa ação de câmara.</li>
          <li><strong>Ferramentas de imagem</strong> — Anexe imagens carregadas ou URLs colados, use variantes de imagem responsivas, remova fundos, recorte moedas e extraia texto de cartões de loja ou certificados.</li>
          <li><strong>PWA/offline</strong> — Instale o Aurearia como PWA para ter navegação ao estilo de aplicação móvel, leitura offline, navegação na coleção em cache e um botão do agente arrastável.</li>
        </ul>

        <h4>Notificações e alertas</h4>
        <ul>
          <li><strong>Notificações na aplicação</strong> — O contador de notificações acompanha a atividade não lida de seguidores, leilões, lista de desejos, marcos de conjuntos e Moeda do dia.</li>
          <li><strong>Pushover</strong> — Adicione uma chave de utilizador em Definições → Conta depois de o administrador configurar o token da aplicação Pushover e o URL público da aplicação.</li>
          <li><strong>Revisão de alertas</strong> — Os alertas de pesquisa da lista de desejos criam execuções com candidatos para rever; os alertas e resumos de leilões ajudam a acompanhar os lotes observados e licitados.</li>
        </ul>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Tipos históricos de moedas</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>As moedas antigas abrangem milhares de anos e muitas civilizações. Estas são as principais categorias:</p>

        <h4>Gregas (c. 600 a.C. – 31 a.C.)</h4>
        <p>Estão entre as primeiras moedas alguma vez cunhadas. As moedas gregas são conhecidas pela beleza artística e pela variedade. Os principais tipos incluem:</p>
        <ul>
          <li><strong>Estáter</strong> — Denominação padrão de ouro ou prata usada nas cidades-estado gregas</li>
          <li><strong>Dracma / Tetradracma</strong> — Moedas de prata; os tetradracmas (4 dracmas) são especialmente apreciados pelos colecionadores pelo tamanho e pelo pormenor artístico</li>
          <li><strong>Óbolo</strong> — Pequena moeda de prata, originalmente 1/6 de dracma</li>
          <li><strong>Hemidracma</strong> — Meio dracma, frequente nas denominações menores</li>
        </ul>
        <p class="my-3 rounded-r-sm border-l-[3px] border-l-gold bg-gold-glow px-[0.85rem] py-[0.6rem] text-body"><strong>Bons pontos de partida:</strong> os tetradracmas atenienses da coruja, os dracmas de Alexandre Magno e os bronzes ptolemaicos são fáceis de encontrar e de reconhecer.</p>

        <h4>República Romana (c. 280 – 27 a.C.)</h4>
        <p>As moedas da República Romana têm desenhos anónimos no início e passam depois a representar figuras políticas e acontecimentos.</p>
        <ul>
          <li><strong>Denário</strong> — A principal moeda de prata de Roma, mais ou menos do tamanho de uma moeda de 10 cêntimos</li>
          <li><strong>Asse / Semis</strong> — Denominações de bronze para o dia a dia</li>
          <li><strong>Vitoriato</strong> — Moeda de prata antiga com o motivo da Vitória</li>
        </ul>

        <h4>Império Romano (27 a.C. – 476 d.C.)</h4>
        <p>A categoria preferida dos colecionadores. As moedas imperiais têm sempre o retrato do imperador no anverso.</p>
        <ul>
          <li><strong>Áureo</strong> — Moeda de ouro (~7,7 g), a denominação mais valiosa</li>
          <li><strong>Denário</strong> — Moeda de prata, a unidade de conta padrão</li>
          <li><strong>Sestércio</strong> — Moeda grande de bronze, apreciada pelos desenhos pormenorizados do reverso</li>
          <li><strong>Antoniniano</strong> — Moeda de prata/bolhão introduzida por Caracala (215 d.C.), progressivamente desvalorizada</li>
          <li><strong>Follis</strong> — Bronze do Baixo Império, comum no século IV</li>
          <li><strong>Sólido</strong> — Moeda de ouro introduzida por Constantino I, que substituiu o áureo</li>
        </ul>
        <p class="my-3 rounded-r-sm border-l-[3px] border-l-gold bg-gold-glow px-[0.85rem] py-[0.6rem] text-body"><strong>Ideal para principiantes:</strong> os bronzes do Baixo Império (século IV) são acessíveis, abundantes e têm retratos de imperadores reconhecíveis. Comece por Constantino I, Constâncio II ou Valentiniano I.</p>

        <h4>Bizantinas (330 – 1453 d.C.)</h4>
        <p>Continuação do Império Romano do Oriente. Destacam-se pelas moedas côncavas (escifatas) e pela iconografia religiosa.</p>
        <ul>
          <li><strong>Sólido / Histamenon</strong> — Moeda de ouro padrão</li>
          <li><strong>Follis</strong> — Moeda de bronze, muitas vezes com Cristo ou o imperador com uma cruz</li>
          <li><strong>Hiperpirão</strong> — Denominação de ouro posterior, muitas vezes côncava</li>
        </ul>

        <h4>Celtas, partas e outras</h4>
        <p>Não ignore as áreas menos colecionadas — as moedas celtas da Gália e da Britânia, os dracmas partos, a prata sassânida, os bronzes da Judeia e as moedas nabateias oferecem oportunidades fascinantes, muitas vezes a preços mais baixos.</p>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Compreender a classificação de moedas</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>A classificação descreve o estado físico da moeda e é o fator que mais pesa no valor. Os graus são padronizados no colecionismo (as abreviaturas internacionais, em inglês, são as que vai encontrar nos anúncios):</p>

        <table>
          <thead>
            <tr><th>Grau</th><th>Abreviatura</th><th>Descrição</th></tr>
          </thead>
          <tbody>
            <tr><td>Poor (mau)</td><td>P / PO</td><td>Mal identificável, desgaste intenso. O tipo e o governante podem ser irreconhecíveis.</td></tr>
            <tr><td>Fair (sofrível)</td><td>FR</td><td>Muito gasta, mas o tipo é identificável. A legenda quase desapareceu.</td></tr>
            <tr><td>About Good (quase regular)</td><td>AG</td><td>Desgaste muito intenso; contorno visível, parte da legenda legível.</td></tr>
            <tr><td>Good (regular)</td><td>G</td><td>Elementos principais do desenho visíveis, mas planos. Legendas parcialmente legíveis.</td></tr>
            <tr><td>Very Good (bom)</td><td>VG</td><td>Elementos principais nítidos com desgaste moderado. Cerca de metade da legenda legível.</td></tr>
            <tr><td>Fine (muito bom)</td><td>F</td><td>Desgaste moderado nos pontos altos. A maior parte da legenda e do desenho visível.</td></tr>
            <tr><td>Very Fine (soberbo)</td><td>VF</td><td>Desgaste ligeiro só nos pontos altos. Legenda completa legível. A maior parte dos pormenores nítidos.</td></tr>
            <tr><td>Extremely Fine (excelente)</td><td>EF / XF</td><td>Desgaste mínimo nos pontos mais altos. Pormenor quase completo. Moeda atraente.</td></tr>
            <tr><td>Almost Uncirculated (quase não circulada)</td><td>AU</td><td>Apenas vestígios de desgaste. Brilho completo nas zonas protegidas.</td></tr>
            <tr><td>Mint State (não circulada)</td><td>MS</td><td>Sem qualquer desgaste. Vai de MS-60 (marcas de saco) a MS-70 (perfeita).</td></tr>
            <tr><td>Superb (flor de cunho)</td><td>FDC</td><td>"Fleur de Coin" — estado de cunhagem perfeito ou quase perfeito. Extremamente raro nas moedas antigas.</td></tr>
          </tbody>
        </table>

        <p class="my-3 rounded-r-sm border-l-[3px] border-l-gold bg-gold-glow px-[0.85rem] py-[0.6rem] text-body"><strong>Nas moedas antigas</strong>, VF é considerado um grau muito respeitável. Muitas moedas antigas em EF ou superior têm prémios significativos. Não espere a perfeição das moedas modernas — moedas com 2000 anos têm carácter!</p>

        <h4>Serviços de classificação</h4>
        <p>Os serviços profissionais de classificação (NGC Ancients, PCGS) autenticam e classificam as moedas, selando-as em cápsulas invioláveis ("slabs"). Isto aumenta o custo, mas dá confiança na autenticidade e no grau. O NGC Ancients é o serviço mais usado para moedas antigas.</p>
        <p>Nem todas as moedas precisam de ser encapsuladas. Muitos colecionadores experientes preferem moedas "soltas" (sem cápsula) pelo prazer de lhes tocar e pelo custo total mais baixo.</p>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Catálogos de referência (RIC, RPC, Sear e outros)</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>Os catálogos de referência atribuem números padronizados aos tipos de moeda, tornando a identificação e a comunicação precisas. Quando vê um número como "RIC 207", refere-se a um tipo específico num volume específico.</p>

        <table>
          <thead>
            <tr><th>Catálogo</th><th>Nome completo</th><th>Âmbito</th></tr>
          </thead>
          <tbody>
            <tr><td><strong>RIC</strong></td><td>Roman Imperial Coinage</td><td>A referência definitiva para as moedas do Império Romano (27 a.C. – 491 d.C.). 10 volumes que abrangem todos os imperadores. O catálogo mais citado para moedas romanas.</td></tr>
            <tr><td><strong>RRC</strong></td><td>Roman Republican Coinage</td><td>Catálogo de Crawford das moedas da República Romana. Citado como "Cr." ou "Crawford".</td></tr>
            <tr><td><strong>RPC</strong></td><td>Roman Provincial Coinage</td><td>Abrange as moedas cunhadas nas províncias romanas (não em Roma). Muitos volumes, também disponíveis online em rpc.ashmus.ox.ac.uk.</td></tr>
            <tr><td><strong>Sear</strong></td><td>Roman Coins and Their Values</td><td>O guia acessível de David Sear, bom para identificação e estimativa de preços. 5 volumes para o Império, com livros separados para as moedas gregas e bizantinas.</td></tr>
            <tr><td><strong>BMC</strong></td><td>British Museum Catalogue</td><td>Referência académica publicada pelo British Museum. Abrange séries gregas, romanas e outras.</td></tr>
            <tr><td><strong>SNG</strong></td><td>Sylloge Nummorum Graecorum</td><td>Catálogo em vários volumes de moedas gregas de coleções de museus de todo o mundo.</td></tr>
            <tr><td><strong>DOC</strong></td><td>Dumbarton Oaks Collection</td><td>A principal referência para moedas bizantinas.</td></tr>
            <tr><td><strong>RSC</strong></td><td>Roman Silver Coins</td><td>Referência da Seaby centrada nas denominações de prata.</td></tr>
          </tbody>
        </table>

        <p class="my-3 rounded-r-sm border-l-[3px] border-l-gold bg-gold-glow px-[0.85rem] py-[0.6rem] text-body"><strong>Para começar:</strong> não precisa de comprar já estes livros caros. Sites como o <a href="https://www.wildwinds.com" target="_blank" rel="noopener">WildWinds</a>, o <a href="https://www.acsearch.info" target="_blank" rel="noopener">ACSearch</a> e o <a href="https://en.numista.com" target="_blank" rel="noopener">Numista</a> disponibilizam bases de dados gratuitas e pesquisáveis com imagens e referências de catálogo.</p>

        <h4>Como ler uma referência de catálogo</h4>
        <p><strong>Exemplo: "RIC VII 162"</strong></p>
        <ul>
          <li><strong>RIC</strong> — O catálogo (Roman Imperial Coinage)</li>
          <li><strong>VII</strong> — Número do volume (época de Constantino I)</li>
          <li><strong>162</strong> — O número do tipo nesse volume</li>
        </ul>
        <p>Indique sempre o número do volume ao citar o RIC, porque a numeração dos tipos recomeça em cada volume.</p>
        <p>Nos dados da moeda, use a secção estruturada <strong>Referências de catálogo</strong> para referências normalizadas (catálogo, volume, número, URI de autoridade). Os campos antigos <code>referenceUrl</code>/<code>referenceText</code> continuam a ser suportados para ligações a anúncios e por compatibilidade.</p>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Ler as legendas (inscrições)</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>As legendas das moedas romanas seguem convenções que, uma vez compreendidas, facilitam muito a identificação. Estão normalmente em latim e leem-se no sentido dos ponteiros do relógio, a partir do canto inferior esquerdo.</p>

        <h4>Elementos comuns no anverso</h4>
        <table>
          <thead>
            <tr><th>Abreviatura</th><th>Significado</th><th>Exemplo</th></tr>
          </thead>
          <tbody>
            <tr><td>IMP</td><td>Imperator (comandante/imperador)</td><td>IMP CAESAR = Imperador César</td></tr>
            <tr><td>CAES / CAESAR</td><td>César (título)</td><td>—</td></tr>
            <tr><td>AVG</td><td>Augusto (título honorífico)</td><td>IMP CAES TRAIANVS AVG</td></tr>
            <tr><td>P M / PONT MAX</td><td>Pontifex Maximus (sumo sacerdote)</td><td>—</td></tr>
            <tr><td>TR P / TRIB POT</td><td>Tribunicia Potestas (poder tribunício) — renovado anualmente, útil para datar</td><td>TR P XV = 15.º ano de poder tribunício</td></tr>
            <tr><td>COS</td><td>Cônsul (com número = qual consulado)</td><td>COS III = terceiro consulado</td></tr>
            <tr><td>P P</td><td>Pater Patriae (Pai da Pátria)</td><td>—</td></tr>
            <tr><td>D N</td><td>Dominus Noster (Nosso Senhor) — Baixo Império</td><td>D N CONSTANTIVS P F AVG</td></tr>
            <tr><td>P F</td><td>Pius Felix (Piedoso e Feliz)</td><td>—</td></tr>
          </tbody>
        </table>

        <h4>Legendas comuns no reverso</h4>
        <table>
          <thead>
            <tr><th>Legenda</th><th>Significado</th></tr>
          </thead>
          <tbody>
            <tr><td>S C (nos bronzes)</td><td>Senatus Consulto — "por decreto do Senado"</td></tr>
            <tr><td>VICTORIA AVG</td><td>Vitória do Imperador</td></tr>
            <tr><td>PAX AVGVSTI</td><td>Paz do Imperador</td></tr>
            <tr><td>CONCORDIA</td><td>Harmonia/Concórdia</td></tr>
            <tr><td>PROVIDENTIA AVG</td><td>Providência do Imperador</td></tr>
            <tr><td>GLORIA EXERCITVS</td><td>Glória do Exército — comum nos bronzes do século IV</td></tr>
            <tr><td>VOT V / X / XX</td><td>Votos por 5/10/20 anos de reinado</td></tr>
            <tr><td>FEL TEMP REPARATIO</td><td>Restauração dos Tempos Felizes — tipo comum da época de Constantino</td></tr>
          </tbody>
        </table>

        <h4>Dicas para ler latim</h4>
        <ul>
          <li><strong>V = U</strong> — O latim não tinha a letra U; AVGVSTVS = Augustus</li>
          <li><strong>I = J</strong> — Não havia a letra J; IVLIA = Júlia</li>
          <li><strong>Ligaduras</strong> — Por vezes as letras partilham traços (AE fundidos em Æ)</li>
          <li><strong>Retrógradas</strong> — As moedas muito antigas têm por vezes letras invertidas</li>
          <li><strong>Legendas gastas</strong> — Use padrões de texto conhecidos e fotografias de referência para completar as falhas</li>
        </ul>

        <p class="my-3 rounded-r-sm border-l-[3px] border-l-gold bg-gold-glow px-[0.85rem] py-[0.6rem] text-body"><strong>Dica de especialista:</strong> a legenda do anverso costuma dizer QUEM (o imperador) e QUANDO (através de títulos como TR P e dos números COS). O reverso transmite a MENSAGEM (propaganda, virtudes, vitórias militares).</p>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Comprar as primeiras moedas</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <h4>Onde comprar</h4>
        <ul>
          <li><strong>Leiloeiras estabelecidas</strong> — CNG (Classical Numismatic Group), Heritage Auctions, Roma Numismatics, Nomos, Leu Numismatik. Estas garantem a autenticidade.</li>
          <li><strong>Comerciantes de confiança</strong> — Procure comerciantes que sejam membros de organizações profissionais (ANA, IAPN, PNG).</li>
          <li><strong>VCoins / MA-Shops</strong> — Mercados online com lojas de comerciantes verificados. As políticas de devolução dão segurança.</li>
          <li><strong>Feiras de moedas</strong> — Ótimas para ver as moedas em mão e conhecer comerciantes. Muitas feiras têm secções de moedas antigas.</li>
        </ul>

        <h4>O que procurar numa moeda</h4>
        <ul>
          <li><strong>Centragem</strong> — Uma cunhagem bem centrada mostra o desenho completo. As moedas descentradas são menos procuradas (exceto se forem muito raras).</li>
          <li><strong>Qualidade do retrato</strong> — Retratos nítidos e pormenorizados são mais atraentes e valiosos.</li>
          <li><strong>Legendas completas</strong> — Legendas legíveis são importantes para a identificação e para o valor.</li>
          <li><strong>Pátina</strong> — A pátina natural (coloração da superfície com o tempo) é desejável. Não limpe moedas antigas.</li>
          <li><strong>Aspeto</strong> — Subjetivo, mas importante. Uma moeda deve ser atraente para o seu grau.</li>
          <li><strong>Proveniência</strong> — O historial documentado de proprietários ("ex-coleção") acrescenta valor e confirma a legitimidade.</li>
        </ul>

        <h4>Orientações de orçamento</h4>
        <table>
          <thead>
            <tr><th>Orçamento</th><th>O que pode esperar</th></tr>
          </thead>
          <tbody>
            <tr><td>10–50 $</td><td>Bronzes do Baixo Império (VG–F), bronzes gregos, lotes por limpar</td></tr>
            <tr><td>50–200 $</td><td>Bronzes romanos melhores (VF), denários comuns (F–VF), bronzes gregos interessantes</td></tr>
            <tr><td>200–500 $</td><td>Bons denários (VF–EF), prata provincial, sestércios (F–VF)</td></tr>
            <tr><td>500–2000 $</td><td>Denários escolhidos (EF+), dracmas gregos de prata, bons sestércios, frações de ouro</td></tr>
            <tr><td>2000+ $</td><td>Áureos/sólidos de ouro, tetradracmas gregos, tipos raros, moedas de grau superior</td></tr>
          </tbody>
        </table>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Cuidados a ter</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <h4>Falsificações</h4>
        <p>A falsificação é um problema na numismática há séculos. As falsificações modernas, sobretudo de certas oficinas no estrangeiro, podem ser muito convincentes. Eis como se proteger:</p>
        <ul>
          <li><strong>Compre a fontes de confiança</strong> — É a regra n.º 1. As leiloeiras e os comerciantes estabelecidos garantem a autenticidade. Se a moeda for falsa, recebe o dinheiro de volta.</li>
          <li><strong>Desconfie do "bom demais para ser verdade"</strong> — Uma moeda rara a preço de saldo é quase de certeza falsa. Se um anúncio no eBay oferece um denário de Júlio César em EF por 50 $, afaste-se.</li>
          <li><strong>Verifique o peso e o diâmetro</strong> — As falsificações têm muitas vezes peso ou dimensões incorretos. Invista numa balança de precisão (0,01 g) e num paquímetro.</li>
          <li><strong>Estude as características dos cunhos</strong> — As moedas autênticas eram cunhadas com cunhos gravados à mão, o que cria variações subtis. As falsificações fundidas parecem muitas vezes "moles" ou têm bolhas de fundição no bordo.</li>
          <li><strong>Procure linhas de molde</strong> — As moedas antigas autênticas eram cunhadas (batidas entre dois cunhos). As falsificações fundidas podem mostrar uma linha à volta do bordo, onde as metades do molde se juntavam.</li>
          <li><strong>Use recursos</strong> — Os <a href="https://www.forumancientcoins.com/fakes/" target="_blank" rel="noopener">Forum Ancient Coins Fake Reports</a> e o grupo Fake Ancient Coins no Facebook são excelentes recursos.</li>
          <li><strong>Encapsule a moeda</strong> — Em compras caras, considere a autenticação NGC ou PCGS.</li>
        </ul>

        <h4>Limpeza</h4>
        <p class="my-3 rounded-r-sm border-l-[3px] border-l-warning bg-status-error-tint px-[0.85rem] py-[0.6rem] text-body text-text-primary">🚫 <strong>Nunca limpe uma moeda antiga.</strong> A limpeza reduz quase sempre o valor da moeda, muitas vezes de forma dramática. A pátina natural que se forma ao longo dos séculos é considerada parte do carácter e da beleza da moeda.</p>
        <ul>
          <li><strong>Nada de escovar, polir ou banhos químicos</strong> — Removem a pátina de forma permanente</li>
          <li><strong>Evite produtos de "limpeza de moedas"</strong> — Foram feitos para moedas modernas e danificam as antigas</li>
          <li><strong>Moedas "retocadas"</strong> — Algumas falsificações ou moedas limpas foram regravadas com uma ferramenta para recuperar pormenor. Procure riscos com padrões pouco naturais nos campos.</li>
          <li><strong>Exceção</strong> — Conservadores profissionais podem estabilizar a "doença do bronze" ou remover depósitos nocivos, mas isso é trabalho especializado, não limpeza.</li>
        </ul>

        <h4>Burlas e armadilhas comuns</h4>
        <ul>
          <li><strong>Lotes de moedas por limpar</strong> — Os lotes de "moedas romanas por limpar" são uma aposta divertida para aprender, mas raramente contêm algo valioso. A maioria serão bronzes do Baixo Império de grau baixo. Não pague demasiado.</li>
          <li><strong>eBay de certas regiões</strong> — Tenha muito cuidado ao comprar a vendedores de regiões conhecidas por produzir falsificações. Prefira vendedores com excelente avaliação e política de devolução.</li>
          <li><strong>"Autenticidade garantida"</strong> — Qualquer pessoa pode escrever isto. Uma garantia vale o que valem a reputação e a política de devolução do vendedor.</li>
          <li><strong>Sobreclassificação</strong> — Os vendedores descrevem por vezes as moedas de forma mais generosa do que merecem. Aprenda a classificar estudando fotografias de referência.</li>
          <li><strong>Atribuições erradas</strong> — Uma moeda identificada como de um imperador pode ser de outro. Confirme sempre a identificação com as referências.</li>
          <li><strong>Questões legais</strong> — Tenha em conta as leis de proteção do património cultural. Compre moedas com proveniência documentada ou a comerciantes que cumpram as regras de importação/exportação. Alguns países restringem a exportação de antiguidades.</li>
        </ul>

        <h4>Resumo dos sinais de alerta</h4>
        <table>
          <thead>
            <tr><th>🚩 Sinal de alerta</th><th>O que significa</th></tr>
          </thead>
          <tbody>
            <tr><td>Preço muito abaixo do valor de mercado</td><td>Provavelmente falsa ou mal descrita</td></tr>
            <tr><td>Sem política de devolução</td><td>O vendedor não confia na autenticidade</td></tr>
            <tr><td>Bolhas de fundição no bordo</td><td>A moeda foi fundida, não cunhada</td></tr>
            <tr><td>Peso ou dimensões errados</td><td>Sinal comum de reprodução moderna</td></tr>
            <tr><td>Superfície brilhante pouco natural</td><td>A moeda foi limpa ou é uma falsificação moderna</td></tr>
            <tr><td>Linha à volta do bordo</td><td>Falsificação fundida num molde de duas partes</td></tr>
            <tr><td>"Qualidade de museu" no eBay</td><td>Linguagem de marketing de vendedores pouco fiáveis</td></tr>
          </tbody>
        </table>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Ligar ferramentas de IA (servidor de ferramentas externo)</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>O servidor de ferramentas externo dá acesso à sua coleção a clientes de IA externos (OpenWebUI, LibreChat, n8n e clientes compatíveis com MCP) através de uma API segura. Os clientes externos podem pesquisar a coleção, obter estatísticas e, opcionalmente, propor alterações através de uma confirmação em duas fases.</p>
        <p>Para o guia técnico completo, consulte <code>docs/external-tool-server.md</code> no repositório.</p>

        <h4>Para administradores</h4>
        <p><strong>Ativar o servidor:</strong> o servidor de ferramentas externo está desativado por predefinição. Para o ativar:</p>
        <ol>
          <li>Aceda a <strong>Administração → IA</strong></li>
          <li>Ative <strong>Ativar servidor de ferramentas externo</strong></li>
          <li>Guarde as definições de IA</li>
        </ol>
        <p>Quando está desativado, todos os pedidos de ferramentas externas falham com <code>503 Service Unavailable</code>. Esta predefinição protege a sua instância até decidir expor a API.</p>
        <p><strong>O que dizer aos utilizadores:</strong> depois de ativado, os utilizadores podem criar chaves de API com âmbito em <strong>Definições → Chaves</strong>. Cada chave pode ser só de leitura (segura para explorar) ou de leitura e escrita (para automação de confiança). As escritas externas ficam registadas no diário de atividade de cada moeda.</p>

        <h4>Para utilizadores</h4>
        <p><strong>Criar uma chave de API:</strong> para ligar um cliente de IA externo à sua coleção:</p>
        <ol>
          <li>Aceda a <strong>Definições → Chaves</strong></li>
          <li>Introduza um nome descritivo (ex.: "OpenWebUI só leitura")</li>
          <li>
            Escolha uma permissão:
            <ul>
              <li><strong>Leitura</strong> (predefinição) — Permite consultas: pesquisa, obter moeda, resumos, moedas mais valiosas</li>
              <li><strong>Leitura/Escrita</strong> — Permite consultas e também propostas e confirmações de alterações</li>
            </ul>
          </li>
          <li>Clique em <strong>Gerar chave</strong></li>
          <li>Copie a chave apresentada (começa por <code>ak_</code>; só é mostrada uma vez)</li>
        </ol>
        <p><strong>Ligar a um cliente:</strong> no seu cliente de IA externo (OpenWebUI, LibreChat, n8n):</p>
        <ol>
          <li>
            Importe o URL OpenAPI:
            <pre>http://your-ancient-coins-host:8080/api/v1/tools/openapi.json</pre>
          </li>
          <li>
            Adicione um cabeçalho personalizado:
            <ul>
              <li><strong>Nome do cabeçalho:</strong> <code>X-API-Key</code></li>
              <li><strong>Valor do cabeçalho:</strong> <code>ak_your_key_here</code></li>
            </ul>
          </li>
          <li>As seis ferramentas da coleção aparecem na sua lista de ferramentas</li>
        </ol>
        <p><strong>Segurança nas escritas:</strong> se criou uma chave de leitura e escrita, as escritas externas exigem dois passos: primeiro propor a alteração (devolve uma pré-visualização) e depois confirmar com <code>confirm=true</code>. Isto evita escritas automáticas acidentais numa conversa casual. Todas as confirmações ficam registadas.</p>

        <h4>Para programadores</h4>
        <p><strong>Caminho base:</strong> <code>/api/v1/tools/*</code></p>
        <p><strong>Autenticação:</strong> envie a chave de API no cabeçalho <code>X-API-Key</code>.</p>
        <p><strong>Ferramentas disponíveis:</strong></p>
        <table>
          <thead>
            <tr><th>Ferramenta</th><th>Permissão</th><th>Finalidade</th></tr>
          </thead>
          <tbody>
            <tr><td><code>search_my_collection</code></td><td>leitura</td><td>Pesquisar moedas por texto</td></tr>
            <tr><td><code>get_coin</code></td><td>leitura</td><td>Obter uma moeda pelo ID</td></tr>
            <tr><td><code>collection_summary</code></td><td>leitura</td><td>Estatísticas agregadas (total de moedas, valor)</td></tr>
            <tr><td><code>top_coins_by_value</code></td><td>leitura</td><td>Moedas mais valiosas, ordenadas pelo valor atual</td></tr>
            <tr><td><code>propose_update</code></td><td>escrita</td><td>Fase 1: propor alterações a campos permitidos</td></tr>
            <tr><td><code>commit_update</code></td><td>escrita</td><td>Fase 2: confirmar com <code>confirm=true</code></td></tr>
          </tbody>
        </table>
        <p><strong>Especificação OpenAPI:</strong> <code>GET /api/v1/tools/openapi.json</code> devolve o esquema completo (sem autenticação; só respeita o interruptor do administrador).</p>
        <p><strong>MCP nativo:</strong> os clientes MCP que suportam Streamable HTTP e um cabeçalho personalizado <code>X-API-Key</code> podem ligar-se diretamente a:</p>
        <pre>http://localhost:8080/api/mcp</pre>
        <p><strong>Segurança:</strong> todas as operações estão limitadas ao utilizador (isolamento por conta). Limite de 50 pedidos/minuto por chave. As escritas externas estão restritas aos mesmos campos permitidos na aplicação (grau, valor atual, notas, etiquetas, referências). Os campos de identidade são rejeitados.</p>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim" :open="route.query.section === 'oidc'">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Configurar o início de sessão OIDC</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <p>O início de sessão OIDC permite aos administradores adicionar o Microsoft Entra ID, o Pocket ID ou outro fornecedor OpenID Connect, mantendo a recuperação por palavra-passe local.</p>
        <p>Para o guia técnico completo, consulte <code>docs/oidc-setup.md</code> no repositório.</p>

        <h4>Antes de começar</h4>
        <ul>
          <li>Mantenha pelo menos uma conta de administrador com credenciais locais utilizáveis, para recuperação.</li>
          <li>Registe no fornecedor os dois URIs de retorno do frontend: início de sessão e associação de contas.</li>
          <li>Use os âmbitos <code>openid profile email</code>, a menos que o fornecedor exija outros.</li>
          <li>Guarde o segredo de cliente apenas nas definições de Administração. Depois de guardado, só pode ser escrito.</li>
        </ul>

        <h4>Microsoft Entra ID</h4>
        <ol>
          <li>Crie um registo de aplicação no centro de administração do Entra.</li>
          <li>Adicione URIs de redirecionamento Web para <code>/auth/oidc/callback/{providerId}</code> e <code>/settings/oidc/link/callback/{providerId}</code>.</li>
          <li>Introduza o ID do inquilino nas definições de Administração; o Aurearia deriva o URL do emissor da Microsoft.</li>
          <li>Copie o ID de cliente da aplicação e o <strong>Valor</strong> do segredo de cliente, não o ID do segredo.</li>
        </ol>

        <h4>Pocket ID</h4>
        <ol>
          <li>Crie um cliente OIDC no Pocket ID.</li>
          <li>Defina o URL do emissor como o URL base do seu Pocket ID.</li>
          <li>Registe os dois caminhos de retorno do frontend como URIs de redirecionamento permitidos.</li>
          <li>Guarde o ID e o segredo de cliente nas definições de Administração e execute o teste do fornecedor.</li>
        </ol>

        <p><strong>Associação de contas:</strong> os utilizadores existentes devem iniciar sessão localmente, abrir Definições → Conta e associar o fornecedor. Endereços de email iguais não são fundidos automaticamente.</p>
      </div>
    </details>

    <details class="group mb-2 overflow-hidden rounded-sm border border-border-subtle open:border-gold-dim">
      <summary class="flex cursor-pointer list-none items-center gap-2 bg-surface px-4 py-3 text-base font-semibold text-text-primary transition-colors before:text-chip before:text-text-muted before:content-['▸'] before:transition-transform hover:bg-card-hover group-open:before:rotate-90 [&::-webkit-details-marker]:hidden">Recursos úteis</summary>
      <div class="data-table overflow-x-auto px-4 py-4 text-base leading-[1.65] text-text-secondary [&_a:hover]:underline [&_code]:text-text-primary [&_h4:first-child]:mt-0 [&_h4]:mt-5 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:text-text-primary [&_li]:mb-[0.35rem] [&_ol]:mb-3 [&_ol]:ml-5 [&_ol]:list-decimal [&_p]:mb-3 [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:whitespace-pre [&_pre]:rounded-sm [&_pre]:border [&_pre]:border-border-subtle [&_pre]:bg-input [&_pre]:p-3 [&_pre]:text-chip [&_pre]:text-text-primary [&_table]:my-3 [&_table]:table-fixed [&_table]:text-body [&_td]:px-[0.65rem] [&_td]:py-2 [&_td]:break-words [&_th]:px-[0.65rem] [&_th]:py-2 [&_ul]:mb-3 [&_ul]:ml-5 [&_ul]:list-disc">
        <h4>Bases de dados e referências online</h4>
        <ul>
          <li><a href="https://www.wildwinds.com" target="_blank" rel="noopener"><strong>WildWinds</strong></a> — Base de dados gratuita de imagens de moedas antigas, organizada por governante e referência de catálogo</li>
          <li><a href="https://en.numista.com" target="_blank" rel="noopener"><strong>Numista</strong></a> — Catálogo colaborativo de moedas com informação pormenorizada por tipo (integrado nesta aplicação)</li>
          <li><a href="https://www.acsearch.info" target="_blank" rel="noopener"><strong>ACSearch</strong></a> — Arquivo de leilões com preços realizados — essencial para avaliações</li>
          <li><a href="https://www.coinarchives.com" target="_blank" rel="noopener"><strong>CoinArchives</strong></a> — Arquivo pesquisável de lotes de leilões anteriores com imagens e preços realizados — ótimo para avaliação e atribuição</li>
          <li><a href="https://rpc.ashmus.ox.ac.uk" target="_blank" rel="noopener"><strong>RPC Online</strong></a> — Base de dados gratuita e pesquisável de moedas provinciais romanas</li>
          <li><a href="https://www.forumancientcoins.com" target="_blank" rel="noopener"><strong>Forum Ancient Coins</strong></a> — Comerciante, artigos educativos e fórum de discussão da comunidade</li>
        </ul>

        <h4>Comunidades</h4>
        <ul>
          <li><a href="https://www.reddit.com/r/AncientCoins/" target="_blank" rel="noopener"><strong>r/AncientCoins</strong></a> — Comunidade ativa no Reddit para ajuda na identificação, compras e discussão</li>
          <li><a href="https://www.cointalk.com/forums/ancient-coins.pair7/" target="_blank" rel="noopener"><strong>CoinTalk Ancient Coins</strong></a> — Fórum numismático de longa data</li>
          <li><strong>Grupos no Facebook</strong> — "Ancient Coins", "Fake Ancient Coins" (educativo) e grupos regionais</li>
        </ul>

        <h4>Livros para principiantes</h4>
        <ul>
          <li><em>Ancient Coin Collecting</em>, de Wayne Sayles — A introdução clássica ao colecionismo</li>
          <li><em>Roman Coins and Their Values</em>, de David Sear — Guia de identificação e preços (5 volumes)</li>
          <li><em>A Dictionary of Ancient Roman Coins</em>, de John Melville Jones — Explica toda a terminologia</li>
          <li><em>Handbook of Greek Coinage</em>, de Oliver Hoover — Se as moedas gregas lhe interessarem</li>
        </ul>
      </div>
    </details>
  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()
</script>
