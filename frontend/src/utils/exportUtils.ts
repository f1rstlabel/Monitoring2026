/**
 * Helper utility for exporting data as CSV or Microsoft Excel (.xls) files.
 */

/**
 * Download data as a UTF-8 CSV file with BOM for Excel compatibility.
 */
export function downloadCSV(
  filename: string,
  headers: string[],
  rows: (string | number | null | undefined)[][]
): void {
  const cleanFilename = filename.endsWith('.csv') ? filename : `${filename}.csv`;
  
  const escapeCell = (val: string | number | null | undefined): string => {
    if (val === null || val === undefined) return '""';
    const str = String(val).replace(/"/g, '""');
    return `"${str}"`;
  };

  const csvRows: string[] = [
    headers.map(escapeCell).join(','),
    ...rows.map(row => row.map(escapeCell).join(','))
  ];

  // Include UTF-8 BOM (\uFEFF) to ensure Microsoft Excel correctly parses UTF-8 encoding
  const csvContent = '\uFEFF' + csvRows.join('\r\n');
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', cleanFilename);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * Download data as a formatted Microsoft Excel (.xls) HTML Spreadsheet document.
 */
export function downloadXLS(
  filename: string,
  title: string,
  headers: string[],
  rows: (string | number | null | undefined)[][]
): void {
  const cleanFilename = filename.endsWith('.xls') ? filename : `${filename}.xls`;
  const exportDate = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  });

  const headerHTML = headers
    .map(
      h =>
        `<th style="background-color: #1e293b; color: #ffffff; font-weight: bold; border: 1px solid #334155; padding: 8px 12px; text-align: left; font-family: sans-serif; font-size: 12px;">${escapeXML(
          h
        )}</th>`
    )
    .join('');

  const rowsHTML = rows
    .map((row, idx) => {
      const bg = idx % 2 === 0 ? '#f8fafc' : '#ffffff';
      const cells = row
        .map(
          cell =>
            `<td style="border: 1px solid #cbd5e1; padding: 6px 12px; font-family: sans-serif; font-size: 11px; color: #0f172a;">${escapeXML(
              cell != null ? String(cell) : ''
            )}</td>`
        )
        .join('');
      return `<tr style="background-color: ${bg};">${cells}</tr>`;
    })
    .join('');

  const excelTemplate = `
    <html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel" xmlns="http://www.w3.org/TR/REC-html40">
    <head>
      <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
      <!--[if gte mso 9]>
      <xml>
        <x:ExcelWorkbook>
          <x:ExcelWorksheets>
            <x:ExcelWorksheet>
              <x:Name>${escapeXML(title.substring(0, 30))}</x:Name>
              <x:WorksheetOptions>
                <x:DisplayGridlines/>
              </x:WorksheetOptions>
            </x:ExcelWorksheet>
          </x:ExcelWorksheets>
        </x:ExcelWorkbook>
      </xml>
      <![endif]-->
    </head>
    <body style="font-family: sans-serif; padding: 10px;">
      <h2 style="color: #1e1b4b; font-size: 16px; font-weight: bold; margin-bottom: 4px;">PEMERINTAH PROVINSI JAWA BARAT — SANOC</h2>
      <h3 style="color: #334155; font-size: 14px; margin-top: 0; margin-bottom: 8px;">${escapeXML(title)}</h3>
      <p style="color: #64748b; font-size: 11px; margin-bottom: 12px;">Tanggal Ekspor: ${exportDate}</p>
      
      <table border="1" style="border-collapse: collapse; width: 100%;">
        <thead>
          <tr>${headerHTML}</tr>
        </thead>
        <tbody>
          ${rowsHTML}
        </tbody>
      </table>
    </body>
    </html>
  `;

  const blob = new Blob([excelTemplate], { type: 'application/vnd.ms-excel;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', cleanFilename);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function escapeXML(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}
