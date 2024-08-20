<tbody id="tbody">
  {{range $val := .}}
  <tr>
    <td>{{$val.Shorty}}</td>
    <td>{{$val.Url}}</td>
    <td>{{$val.Expired}}</td>
    {{end}}
</tbody>
